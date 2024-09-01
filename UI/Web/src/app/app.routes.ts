import { Routes } from '@angular/router';
import {AuthGuard} from "./_guards/auth.guard";

export const routes: Routes = [
  {
    path: '',
    canActivate: [AuthGuard],
    runGuardsAndResolvers: 'always',
    children: [
      {
        path: 'home',
        loadChildren: () => import('./_routes/dashboard.routes').then(m => m.routes)
      },
      {
        path: 'page',
        loadChildren: () => import('./_routes/page.routes').then(m => m.routes)
      },
      {path: '', pathMatch: 'full', redirectTo: 'home'},
    ]
  },
  {
    path: 'login',
    loadChildren: () => import('./_routes/registration.routes').then(m => m.routes)
  }
];
