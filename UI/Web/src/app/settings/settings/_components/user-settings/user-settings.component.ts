import {Component, inject} from '@angular/core';
import {AccountService} from "../../../../_services/account.service";
import {hasPermission, Perm, User, UserDto} from "../../../../_models/user";
import {Clipboard} from "@angular/cdk/clipboard";
import {FormsModule} from "@angular/forms";
import {ToastService} from '../../../../_services/toast.service';
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {TableComponent} from "../../../../shared/_component/table/table.component";
import {NgbTooltip} from "@ng-bootstrap/ng-bootstrap";
import {ModalService} from "../../../../_services/modal.service";

@Component({
  selector: 'app-user-settings',
  imports: [
    FormsModule,
    TranslocoDirective,
    TableComponent,
    NgbTooltip
  ],
  templateUrl: './user-settings.component.html',
  styleUrl: './user-settings.component.scss'
})
export class UserSettingsComponent {

  private readonly modalService = inject(ModalService);

  users: UserDto[] = []
  authUser: User | null = null;

  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;

  constructor(private accountService: AccountService,
              private toastService: ToastService,
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
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  copyApiKey() {
    this.clipBoard.copy(this.authUser!.apiKey)
  }

  async resetApiKey() {
    if (!await this.modalService.confirm({
      question: translate("settings.users.confirm-reset-api-key")
    })) {
      return;
    }

    this.accountService.refreshApiKey().subscribe({
      next: res => {
        this.clipBoard.copy(res.ApiKey)
        this.toastService.successLoco("settings.users.toasts.reset-api-key.success");
      },
      error: err => {
        this.toastService.errorLoco("settings.users.toasts.reset-api-key.error", {}, {msg: err.error.message});
      }
    })
  }

  async resetPassword(user: UserDto) {
    if (!await this.modalService.confirm({
      question: translate("settings.users.confirm-reset-password-password", {name: user.name})
    })) {
      return;
    }

    this.accountService.generateReset(user.id).subscribe({
      next: reset => {
        this.clipBoard.copy(`/login/reset?key=${reset.Key}`)
        this.toastService.successLoco("settings.users.toasts.reset-password.success");
      },
      error: err => {
        this.toastService.errorLoco("settings.users.toasts.reset-password.error", {}, {msg: err.error.message});
      }
    })
  }

  async deleteUser(user: UserDto) {
    if (!await this.modalService.confirm({
      question: translate("settings.user.confirm-delete", {name: user.name})
    })) {
      return;
    }

    this.accountService.delete(user.id).subscribe({
      next: _ => {
        this.users = this.users.filter(dto => dto.id !== user.id)
        this.toastService.successLoco("settings.users.toasts.delete.success", {name: user.name});
      },
      error: err => {
        this.toastService.errorLoco("settings.users.toasts.delete.error",
          {name: user.name}, {msg: err.error.message});
      }
    })
  }

  emptyUserPresent() {
    return this.users.find(user => user.id === 0) !== undefined;
  }

  trackBy(idx: number, user: UserDto) {
    return `${user.id}`
  }
}
