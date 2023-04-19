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

import {Component, Injectable, OnInit } from '@angular/core';
import {UntypedFormGroup} from '@angular/forms';
import {SegmentService} from 'ngx-segment-analytics';
import {environment} from "../../environments/environment";
import {Features, ApplicationService, FeaturesFEATURE} from "../app.service";
import {ActivatedRoute, ActivatedRouteSnapshot, Resolve, RouterStateSnapshot } from '@angular/router';
import {Observable} from 'rxjs';

@Component({
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {
  public error = '';
  public updated = '';
  public emailForm!: UntypedFormGroup;
  public submitted!: boolean;
  public features!: Array<FeaturesFEATURE>;
  public FeaturesFEATURE = FeaturesFEATURE;

  constructor(
    private readonly route: ActivatedRoute,
  ) {}

  ngOnInit(): void {
    this.route.data.subscribe(
      d => {
        this.features = d[0].features;
      },
      err => {
        this.error = err;
      });
  }

  get f(): any {
    return this.emailForm.controls;
  }
}

@Injectable({providedIn: 'root'})
export class FeatureResolver implements Resolve<Features> {
  constructor(private readonly appService: ApplicationService) {}

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<Features> {
    return this.appService.getFeatures('');
  }
}