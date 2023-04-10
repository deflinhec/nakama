// Copyright 2020 The Nakama Authors
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

import {NgModule} from '@angular/core';
import {Routes, RouterModule} from '@angular/router';
import {HomeComponent} from './home/home.component';
import {ResetPasswordComponent, ResetGuard} from './reset-password/reset-password.component';
import {ForgotPasswordComponent} from './forgot-password/forgot-password.component';
import {EmailVerificationLinkComponent} from './email-verification/link/link.component';
import {EmailVerificationCodeComponent} from './email-verification/code/code.component';

const routes: Routes = [
  {path: '', 
      component: HomeComponent},
  {path: 'reset-password', 
      component: ResetPasswordComponent, 
      canActivate: [ResetGuard]},
  {path: 'forgot-password', 
      component: ForgotPasswordComponent},
  {path: 'email-verification', children: [
      {path: 'link', 
        component: EmailVerificationLinkComponent},
      {path: 'code', 
          component: EmailVerificationCodeComponent},
    ],
  },
  
  // Fallback redirect.
  {path: '**', redirectTo: ''}
];

@NgModule({
  imports: [
    RouterModule.forRoot(routes, {useHash: true}),
    // RouterModule.forRoot(routes, { useHash: true, enableTracing: true }), // TODO debugging purposes only
  ],
  exports: [RouterModule]
})
export class AppRoutingModule { }
