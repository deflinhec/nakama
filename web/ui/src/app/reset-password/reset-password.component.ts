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
import {UntypedFormBuilder, UntypedFormGroup, Validators} from '@angular/forms';
import {ActivatedRoute, ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {ApplicationService, VerifyPasswordRenewalRequest} from '../app.service';
import {SegmentService} from 'ngx-segment-analytics';
import {environment} from "../../environments/environment";

@Component({
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.scss']
})
export class ResetPasswordComponent implements OnInit {
  public error = '';
  public updated = '';
  public resetForm!: UntypedFormGroup;
  public submitted!: boolean;

  constructor(
    private segment: SegmentService,
    private readonly formBuilder: UntypedFormBuilder,
    private route: ActivatedRoute, 
    private readonly appService: ApplicationService,
  ) {}

  ngOnInit(): void {
    if (!environment.nt) {
      this.segment.page('/reset-password');
    }
    this.resetForm = this.formBuilder.group({
      newPassword: ['', Validators.compose([
        Validators.required,
        Validators.minLength(8),
      ])],
      confirmPassword: ['', Validators.compose([
        Validators.required, 
        Validators.minLength(8),
      ])],
    });
  }

  onSubmit(): void {
    this.submitted = true;
    this.error = '';
    if (this.f.invalid) {
      return;
    }
    const token = this.route.snapshot.queryParamMap.get('token');
    if (this.f.newPassword.value !== this.f.confirmPassword.value) {
      this.error = 'Passwords do not match.';
      this.submitted = false;
      return;
    }
    const body : VerifyPasswordRenewalRequest = {
      password: this.f.newPassword.value,
      token: token,
    };
    this.appService.verifyPasswordRenewal('', body)
      .subscribe(d => {
        this.updated = 'Password reset successful.';
        this.error = '';
      }, err => {
        this.updated = '';
        this.error = err.error.message;
        this.submitted = false;
      });
  }

  get f(): any {
    return this.resetForm.controls;
  }
}

@Injectable({providedIn: 'root'})
export class ResetGuard implements CanActivate {
  constructor(
    private readonly router: Router
  ) {}

  canActivate(next: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean {
    const token = next.queryParamMap.get('token');
    if (token == null) {
      const _ = this.router.navigate(['/']);
      return false;
    }
    return true;
  }
}