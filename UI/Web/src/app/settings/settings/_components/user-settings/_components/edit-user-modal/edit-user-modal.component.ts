import {ChangeDetectionStrategy, Component, computed, effect, inject, model, OnInit, signal} from '@angular/core';
import {AllPerms, Perm, roles, UserDto} from "../../../../../../_models/user";
import {NgbActiveModal} from "@ng-bootstrap/ng-bootstrap";
import {AccountService} from "../../../../../../_services/account.service";
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {SettingsItemComponent} from "../../../../../../shared/form/settings-item/settings-item.component";
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {TypeaheadComponent, TypeaheadSettings} from "../../../../../../type-ahead/typeahead.component";
import {PermNamePipe} from "../../../../../../_pipes/perm-name.pipe";
import {of} from "rxjs";
import {DefaultValuePipe} from "../../../../../../_pipes/default-value.pipe";
import {ToastService} from "../../../../../../_services/toast.service";

@Component({
  selector: 'app-edit-user-modal',
  imports: [
    TranslocoDirective,
    SettingsItemComponent,
    ReactiveFormsModule,
    TypeaheadComponent,
    PermNamePipe,
    DefaultValuePipe
  ],
  templateUrl: './edit-user-modal.component.html',
  styleUrl: './edit-user-modal.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class EditUserModalComponent implements OnInit {

  private readonly toastService = inject(ToastService);
  private readonly userService = inject(AccountService);
  private readonly modal = inject(NgbActiveModal);
  private readonly permPipe = inject(PermNamePipe);

  user = model.required<UserDto>();

  userName = computed(() => {
    const user = this.user();
    if (user.name) return user.name;

    return translate('edit-user-modal.someone');
  });
  selectedPerms = signal<Perm[]>([]);


  userForm = new FormGroup({});

  constructor() {
    effect(() => this.selectedPerms.set(roles(this.user())));
  }

  ngOnInit() {
    const user = this.user();

    this.userForm.addControl('id', new FormControl(user.id));
    this.userForm.addControl('name', new FormControl(user.name, [Validators.required]));
    this.userForm.addControl('email', new FormControl(user.email));

  }

  rolesTypeaheadSettings(): TypeaheadSettings<Perm> {
    const settings = new TypeaheadSettings<Perm>();
    settings.multiple = true;
    settings.minCharacters = 0;
    settings.id = 'role-typeahead';

    settings.fetchFn = (f) =>
      of(AllPerms.filter(p => this.permPipe.transform(p).includes(f)));
    settings.savedData = roles(this.user());


    return settings;
  }

  updatePerms(perms: Perm[] | Perm) {
    this.selectedPerms.set(perms as Perm[]);
  }

  close() {
    this.modal.close();
  }

  packData() {
    const data = this.userForm.value as UserDto;

    data.id = data.id === -1 ? 0 : data.id;
    data.permissions = this.selectedPerms().reduce((old, cur) => old | cur, 0);
    return data;
  }

  save() {
    const user = this.packData();

    this.userService.updateOrCreate(user).subscribe({
      next: () => this.toastService.infoLoco("settings.users.toasts.updated.success", {name: this.userName()}),
      error: (err) => this.toastService.errorLoco("settings.users.toasts.update.error",
        {name: this.userName()}, {msg: err.error.message})
    }).add(() => this.close());
  }

}
