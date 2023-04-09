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

import {Component, OnInit } from '@angular/core';
import {UntypedFormBuilder, UntypedFormGroup, Validators} from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {ApplicationService, SendEmailVerificationRequest} from '../../app.service';
import {SegmentService} from 'ngx-segment-analytics';
import {environment} from "../../../environments/environment";

@Component({
  templateUrl: './code.component.html',
  styleUrls: ['./code.component.scss']
})
export class EmailVerificationCodeComponent implements OnInit {
  public error = '';
  public updated = '';
  public emailForm!: UntypedFormGroup;
  public submitted!: boolean;

  constructor(
    private segment: SegmentService,
    private readonly formBuilder: UntypedFormBuilder,
    private route: ActivatedRoute, 
    private readonly appService: ApplicationService,
  ) {}

  ngOnInit(): void {
    if (!environment.nt) {
      this.segment.page('/email-verification/code');
    }
    this.emailForm = this.formBuilder.group({
      email: ['', Validators.compose([
        Validators.required,
        Validators.email,
      ])],
    });
  }

  onSubmit(): void {
    this.submitted = true;
    this.error = '';
    if (this.f.invalid) {
      return;
    }
    const body : SendEmailVerificationRequest = {
      email: this.f.email.value,
    };
    this.appService.sendEmailVerificationCode('', body)
      .subscribe(d => {
        this.updated = 'Email have been sent.';
        this.error = '';
      }, err => {
        this.updated = '';
        this.error = err.error.message;
        this.submitted = false;
      });
  }

  get f(): any {
    return this.emailForm.controls;
  }
}
