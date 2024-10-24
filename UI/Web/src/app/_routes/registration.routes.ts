import {Routes} from "@angular/router";
import {UserLoginComponent} from "../registration/user-login/user-login.component";
import {ResetComponent} from "../registration/reset/reset.component";

export const routes: Routes = [
  {
    path: '',
    component: UserLoginComponent
  },
  {
    path: 'reset',
    component: ResetComponent
  }
]
