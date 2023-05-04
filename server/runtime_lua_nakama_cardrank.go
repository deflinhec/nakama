// Copyright 2023 Deflinhec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"

	lua "github.com/heroiclabs/nakama/v3/internal/gopher-lua"

	"github.com/cardrank/cardrank"
)

// @group cardrank
// @summary Evaulate card ranks.
// @param pockets(type=table) A list of pockets which contain card indexes.
// @param board(type=table) A list of public card indexes.
// @param rule(type=string) The unique identifier of evaluation rule.
// @param pivot(type=bool) Whether the results has a winner.
// @return results(table) Evaulation results.
// @return error(error) An optional error value if an error occurred.
func (n *RuntimeLuaNakamaModule) cardrankEvaluate(l *lua.LState) int {
	table := l.CheckTable(1)
	pockets := [][]cardrank.Card{}
	for i := 1; i <= table.MaxN(); i++ {
		if indexes, ok := table.RawGetInt(i).(*lua.LTable); ok {
			pocket := make([]cardrank.Card, 0, indexes.MaxN())
			for j := 1; j <= indexes.MaxN(); j++ {
				value := RuntimeLuaConvertLuaValue(indexes.RawGetInt(j))
				if i64, ok := value.(int64); ok {
					card := cardrank.FromIndex(int(i64))
					if card == cardrank.InvalidCard {
						l.ArgError(1, "expects card must be a valid index")
						return 0
					}
					pocket = append(pocket, card)
				} else {
					l.ArgError(1, "expects card must be an integer")
					return 0
				}
			}
			pockets = append(pockets, pocket)
		} else {
			l.ArgError(1, "expects pocket must be a table")
			return 0
		}
	}

	table = l.CheckTable(2)
	board := []cardrank.Card{}
	for i := 1; i <= table.MaxN(); i++ {
		value := RuntimeLuaConvertLuaValue(table.RawGetInt(i))
		if i64, ok := value.(int64); ok {
			card := cardrank.FromIndex(int(i64))
			if card == cardrank.InvalidCard {
				l.ArgError(2, "expects card must be a valid index")
				return 0
			}
			board = append(board, card)
		} else {
			l.ArgError(2, "expects card must be an integer")
			return 0
		}
	}

	id := l.OptString(3, cardrank.Holdem.Id())
	rule, err := cardrank.IdToType(id)
	if err != nil {
		l.ArgError(3, "expects rule must be valid, "+err.Error())
		return 0
	}

	evs := rule.EvalPockets(pockets, board)
	order, pivot := cardrank.Order(evs, false)

	results := l.CreateTable(len(evs), 0)
	for i, ev := range evs {
		desc := ev.Desc(false)
		item := l.CreateTable(0, 4)
		item.RawSetString("rank", RuntimeLuaConvertValue(l, order[i]+1))
		item.RawSetString("best", RuntimeLuaConvertCardrankIndexes(l, desc.Best))
		item.RawSetString("unused", RuntimeLuaConvertCardrankIndexes(l, desc.Unused))
		info := l.CreateTable(0, 4)
		info.RawSetString("type", lua.LString(desc.Rank.Name()))
		info.RawSetString("rule", lua.LString(desc.Type.Name()))
		info.RawSetString("best", lua.LString(fmt.Sprintf("%c",
			cardrank.CardFormatter(desc.Best))))
		info.RawSetString("unused", lua.LString(fmt.Sprintf("%c",
			cardrank.CardFormatter(desc.Unused))))
		item.RawSetString("desc", info)
		results.RawSetInt(i+1, item)
	}
	l.Push(lua.LBool(pivot == 1))
	l.Push(results)
	return 2
}

// @group cardrank
// @summary Card formatter.
// @param cards(type=table) A list of card indexes.
// @return results(type=string) Cards vaule representaion in text.
// @return error(error) An optional error value if an error occurred.
func (n *RuntimeLuaNakamaModule) cardrankFormat(l *lua.LState) int {
	pattern := l.CheckString(1)
	table := l.CheckTable(2)
	cards := []cardrank.Card{}
	for i := 1; i <= table.MaxN(); i++ {
		value := RuntimeLuaConvertLuaValue(table.RawGetInt(i))
		if i64, ok := value.(int64); ok {
			card := cardrank.FromIndex(int(i64))
			if card == cardrank.InvalidCard {
				l.ArgError(1, fmt.Sprintf("invalid card value %d at index %d", i64, i))
				return 0
			}
			cards = append(cards, card)
		} else {
			l.ArgError(1, fmt.Sprintf("invalid card value %v at index %d", i, value))
			return 0
		}
	}
	l.Push(lua.LString(fmt.Sprintf(pattern, cardrank.CardFormatter(cards))))
	return 1
}

func RuntimeLuaConvertCardrankIndexes(l *lua.LState, cards []cardrank.Card) lua.LValue {
	table := l.CreateTable(len(cards), 0)
	for i, card := range cards {
		table.RawSetInt(i+1, lua.LNumber(card.Index()))
	}
	return table
}
