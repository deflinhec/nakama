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

	"github.com/cardrank/cardrank"
	"github.com/dop251/goja"
)

// @group cardrank
// @summary Evaulate card ranks.
// @param pockets(type=table) A list of pockets which contain card indexes.
// @param board(type=table) A list of public card indexes.
// @param rule(type=string) The unique identifier of evaluation rule.
// @param pivot(type=bool) Whether the results has a winner.
// @return results(table) Evaulation results.
// @return error(error) An optional error value if an error occurred.
func (n *runtimeJavascriptNakamaModule) cardrankEvaluate(r *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(f goja.FunctionCall) goja.Value {
		pocketsIn, ok := f.Argument(0).Export().([]interface{})
		if !ok {
			panic(r.NewTypeError("expects pockets must be an array"))
		}
		pockets := [][]cardrank.Card{}
		for i, pocketIn := range pocketsIn {
			indexesIn, ok := pocketIn.([]interface{})
			if !ok {
				panic(r.NewTypeError("expects pocket must be an array"))
			}
			pockets = append(pockets, []cardrank.Card{})
			for _, indexIn := range indexesIn {
				i64, ok := indexIn.(int64)
				if !ok {
					panic(r.NewTypeError("expects pocket index must be an array"))
				}
				card := cardrank.FromIndex(int(i64))
				if card == cardrank.InvalidCard {
					panic(r.NewTypeError("expects card must be a valid index"))
				}
				pockets[i] = append(pockets[i], card)
			}
		}

		boardIn, ok := f.Argument(1).Export().([]interface{})
		if !ok {
			panic(r.NewTypeError("expects board must be an array"))
		}
		board := []cardrank.Card{}
		for _, indexIn := range boardIn {
			i64, ok := indexIn.(int64)
			if !ok {
				panic(r.NewTypeError("expects card must be an integer"))
			}
			card := cardrank.FromIndex(int(i64))
			if card == cardrank.InvalidCard {
				panic(r.NewTypeError("expects card must be a valid index"))
			}
			board = append(board, card)
		}

		rule := cardrank.Holdem
		if f.Argument(2) != goja.Undefined() && f.Argument(3) != goja.Null() {
			var err error
			id := getJsString(r, f.Argument(2))
			if rule, err = cardrank.IdToType(id); err != nil {
				panic(r.NewTypeError("expects rule must be valid, " + err.Error()))
			}
		}

		evs := rule.EvalPockets(pockets, board)
		order, pivot := cardrank.Order(evs, false)

		results := make([]interface{}, 0, len(evs))
		for i, ev := range evs {
			desc := ev.Desc(false)
			result := map[string]interface{}{
				"rank":   order[i] + 1,
				"best":   getJsCardrankIndexes(r, desc.Best),
				"unused": getJsCardrankIndexes(r, desc.Unused),
				"desc": map[string]interface{}{
					"type":   desc.Rank.Name(),
					"rule":   desc.Type.Name(),
					"best":   fmt.Sprintf("%c", cardrank.CardFormatter(desc.Best)),
					"unused": fmt.Sprintf("%c", cardrank.CardFormatter(desc.Unused)),
				},
			}
			results = append(results, result)
		}

		return r.ToValue(map[string]interface{}{
			"pivot":   pivot == 1,
			"results": results,
		})
	}
}

func (n *runtimeJavascriptNakamaModule) cardrankFormat(r *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(f goja.FunctionCall) goja.Value {
		pattern := getJsString(r, f.Argument(0))
		cardsIn, ok := f.Argument(1).Export().([]interface{})
		if !ok {
			panic(r.NewTypeError("expects board must be an array"))
		}
		cards := []cardrank.Card{}
		for _, indexIn := range cardsIn {
			index, ok := indexIn.(int64)
			if !ok {
				panic(r.NewTypeError("expects board index must be an array"))
			}
			card := cardrank.FromIndex(int(index))
			if card == cardrank.InvalidCard {
				panic(r.NewTypeError("expects board index must be an array"))
			}
			cards = append(cards, card)
		}
		return r.ToValue(fmt.Sprintf(pattern, cardrank.CardFormatter(cards)))
	}
}

func getJsCardrankIndexes(r *goja.Runtime, cards []cardrank.Card) []interface{} {
	indexes := make([]interface{}, len(cards))
	for i, card := range cards {
		indexes[i] = int(card.Index())
	}
	return indexes
}
