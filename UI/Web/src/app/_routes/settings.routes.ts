import {Routes} from "@angular/router";
import {SettingsComponent} from "../settings/settings/settings.component";
import {PageWizardComponent} from "../settings/settings/_wizard/page-wizard/page-wizard.component";

export const routes: Routes = [
  {
    path: '',
    component: SettingsComponent,
  },
  {
    path: 'wizard',
    children: [
      {
        path: 'page',
        component: PageWizardComponent,
      }
    ]
  }
]
