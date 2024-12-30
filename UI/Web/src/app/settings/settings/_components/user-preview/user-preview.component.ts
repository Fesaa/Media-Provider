import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {hasPermission, Perm, User, UserDto} from '../../../../_models/user';
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {AccountService} from "../../../../_services/account.service";
import {PermissionSettingsComponent} from "../permission-settings/permission-settings.component";
import {NgIcon} from "@ng-icons/core";
import {DialogService} from "../../../../_services/dialog.service";
import {ToastrService} from "ngx-toastr";
import {Tooltip} from "primeng/tooltip";
import {Clipboard} from "@angular/cdk/clipboard";

@Component({
    selector: 'app-user-preview',
  imports: [
    PermissionSettingsComponent,
    ReactiveFormsModule,
    NgIcon,
    Tooltip,
  ],
    templateUrl: './user-preview.component.html',
    styleUrl: './user-preview.component.css'
})
export class UserPreviewComponent implements OnInit {

  @Output() deleteUserEmitter: EventEmitter<number> = new EventEmitter();
  @Output() updateIdEmitter: EventEmitter<{old: number, new: number}> = new EventEmitter();
  @Input({required: true}) user!: UserDto;
  @Input() delete: boolean = false;
  form!: FormGroup;

  edit: boolean = false;
  authUser!: User;

  constructor(
    private fb: FormBuilder,
    private accountService: AccountService,
    private dialogService: DialogService,
    private toastR: ToastrService,
    private clipBoard: Clipboard,
  ) {
  }

  refreshApiKey() {
    if (this.authUser.id !== this.user.id) {
      return;
    }

    this.accountService.refreshApiKey().subscribe({
      next: data => {
        this.toastR.success("Refreshed API key");
      },
      error: err => {
        this.toastR.error("Failed to refresh API key", err.message);
      }
  })
  }

  copyAuth() {
    if (this.authUser.id === this.user.id) {
      this.clipBoard.copy(this.authUser.apiKey);
      this.toastR.success('Removing in 1m', 'Api Key copied to clipboard');
      setTimeout(() => {
        this.clipBoard.copy('')
      }, 60 * 1000);
    }
  }

  ngOnInit(): void {
    this.accountService.currentUser$.subscribe((user) => {
      if (user) {
        this.authUser = user;
      }
    })
    this.form = this.fb.group({
      name: this.fb.control(this.user.name, Validators.required),
      permissions: this.fb.control(this.valueToPermissions(), Validators.required)
    })
  }

  async submit() {
    const dto: UserDto = {
      id: this.user.id,
      name: this.form.value.name,
      permissions: this.permissionsToValue()
    }

    if (dto.permissions !== this.user.permissions) {
      if (!(await this.dialogService.openDialog(`You are changing ${dto.name}'s permissions, are you sure?`))) {
        return;
      }
    }

    this.accountService.update(dto).subscribe({
      next: (id) => {
        this.updateIdEmitter.emit({old: this.user.id, new: id});

        this.user = dto;
        this.user.id = id;
        this.toastR.success(`${dto.name} updated successfully`, 'Success');
      },
      error: (err) => {
        this.toastR.error(err.error.error, 'Error');
      },
      complete: () => {
        this.edit = false;
      }
    })

  }

  async deleteUser() {
    if (!this.delete) {
      this.toastR.warning('You cannot delete this account.', "Warning!");
      return;
    }

    if (this.user.id === 0) {
      this.deleteUserEmitter.emit(this.user.id);
      return;
    }

    if (!(await this.dialogService.openDialog(`You are deleting ${this.user.name}'s account, are you sure?`))) {
      return;
    }


    this.accountService.delete(this.user.id).subscribe({
      next: () => {
        this.deleteUserEmitter.emit(this.user.id);
        this.toastR.success(`Deleted ${this.user.name} successfully`, 'Success');
      },
      error: (err) => {
        this.toastR.error(err.error.error, 'Error');
      }
    })
  }

  permissionsToValue() {
    let val = 0;
    for (const perm of (this.form.value.permissions as Perm[])) {
      val |= perm;
    }
    return val;
  }

  valueToPermissions() {
    return Object.values(Perm)
      .filter(v => typeof v === 'number')
      .filter(v => hasPermission(this.user, v));
  }

  toggleEdit() {
    this.edit = !this.edit;
  }

  async resetPassword() {
    if (!this.delete) {
      this.toastR.warning('You cannot reset the password for this account.', "Warning!");
      return;
    }

    if (!(await this.dialogService.openDialog(`Are you sure you want to generate a password reset for ${this.user.name}?`))) {
      return;
    }

    this.accountService.generateReset(this.user.id).subscribe({
      next: () => {
        this.toastR.success(`${this.user.name} generated reset successfully. View server logs for the key`, 'Success');
      },
      error: (err) => {
        this.toastR.error(err.error.error, 'Error');
      }
    });
  }

  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
}
