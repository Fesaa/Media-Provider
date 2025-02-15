import {Component} from '@angular/core';
import {AccountService} from "../../../../_services/account.service";
import {hasPermission, Perm, permissionNames, permissionValues, roles, User, UserDto} from "../../../../_models/user";
import {TableModule} from "primeng/table";
import {Button} from "primeng/button";
import {DialogService} from "../../../../_services/dialog.service";
import {Tooltip} from "primeng/tooltip";
import {Clipboard} from "@angular/cdk/clipboard";
import {Dialog} from "primeng/dialog";
import {InputText} from "primeng/inputtext";
import {FormsModule} from "@angular/forms";
import {MultiSelect} from "primeng/multiselect";
import {FloatLabel} from "primeng/floatlabel";
import {MessageService} from '../../../../_services/message.service';

@Component({
  selector: 'app-user-settings',
  imports: [
    TableModule,
    Button,
    Tooltip,
    Dialog,
    InputText,
    FormsModule,
    MultiSelect,
    FloatLabel
  ],
  templateUrl: './user-settings.component.html',
  styleUrl: './user-settings.component.css'
})
export class UserSettingsComponent {

  users: UserDto[] = []
  authUser: User | null = null;
  loading: boolean = true;

  showEditUserModal: boolean = false;
  editingUser: UserDto | null = null;
  editPermissions: Perm[] = [];

  possiblePermissions: { label: string, value: Perm }[] = [
    {value: Perm.WriteUser, label: 'Write User'},
    {value: Perm.DeleteUser, label: 'Delete User'},
    {value: Perm.WritePage, label: 'Write Page'},
    {value: Perm.DeletePage, label: 'Delete User'},
    {value: Perm.WriteConfig, label: 'Write Config'},
  ];
  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
  protected readonly roles = roles;
  protected readonly permissionValues = permissionValues;
  protected readonly permissionNames = permissionNames;

  constructor(private accountService: AccountService,
              private msgService: MessageService,
              private dialogService: DialogService,
              private clipBoard: Clipboard,
  ) {
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.authUser = user;
      }
    })

    this.accountService.all().subscribe({
      next: users => {
        this.users = users;
        this.loading = false;
      },
      error: err => {
        this.msgService.error("Unable to load all users", err.error.message)
      }
    })
  }

  permissionsToValue() {
    let val = 0;
    for (const perm of (this.editPermissions as Perm[])) {
      val |= perm;
    }
    return val;
  }

  saveEdit() {
    this.showEditUserModal = false;
    if (this.editingUser == null) {
      return;
    }

    this.editingUser.permissions = this.permissionsToValue()
    this.accountService.updateOrCreate(this.editingUser).subscribe({
      next: dto => {
        this.users = this.users.filter(user => user.id !== dto.id)
        this.users.push(dto)
        this.msgService.info("Update User");
      },
      error: err => {
        this.msgService.error(err.error.message, "Unable to update User");
      }
    })
  }

  editUser(user: UserDto) {
    this.editPermissions = roles(user);
    this.editingUser = user;
    this.showEditUserModal = true;
  }

  newUser() {
    this.editUser({
      id: 0,
      name: '',
      canDelete: true,
      permissions: -1,
    });
  }

  copyApiKey() {
    this.clipBoard.copy(this.authUser!.apiKey)
  }

  async resetApiKey() {
    if (!await this.dialogService.openDialog(`Reset your ApiKey`)) {
      return;
    }

    this.accountService.refreshApiKey().subscribe({
      next: res => {
        this.clipBoard.copy(res.ApiKey)
        this.msgService.success("Reset your ApiKey", "Key has been copied to clipboard")
      },
      error: err => {
        this.msgService.error("Unable to refresh api key", err.error.message)
      }
    })
  }

  async resetPassword(user: UserDto) {
    if (!await this.dialogService.openDialog(`Reset password ${user.name}`)) {
      return;
    }

    this.accountService.generateReset(user.id).subscribe({
      next: reset => {
        this.clipBoard.copy(`/login/reset?key=${reset.Key}`)
        this.msgService.success("Reset generated", "The reset link has been copied to your clipboard. A copy may be found in your server logs")
      },
      error: err => {
        this.msgService.error("Failed to generate reset link", err.error.message);
      }
    })
  }

  async deleteUser(user: UserDto) {
    if (!await this.dialogService.openDialog(`Delete ${user.name}`)) {
      return;
    }

    this.accountService.delete(user.id).subscribe({
      next: _ => {
        this.users = this.users.filter(dto => dto.id !== user.id)
        this.msgService.success(`${user.name} has been deleted`)
      },
      error: err => {
        this.msgService.error("Unable to delete user", err.error.message)
      }
    })
  }

  emptyUserPresent() {
    return this.users.find(user => user.id === 0) !== undefined;
  }
}
