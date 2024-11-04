import { Component } from '@angular/core';
import {AccountService} from "../../../../_services/account.service";
import {hasPermission, Perm, User, UserDto} from "../../../../_models/user";
import {UserPreviewComponent} from "../user-preview/user-preview.component";
import {NgIcon} from "@ng-icons/core";

@Component({
  selector: 'app-user-settings',
  standalone: true,
  imports: [
    UserPreviewComponent,
    NgIcon
  ],
  templateUrl: './user-settings.component.html',
  styleUrl: './user-settings.component.css'
})
export class UserSettingsComponent {

  users: UserDto[] = []
  authUser: User | null = null;

  constructor(private accountService: AccountService,) {
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.authUser = user;
      }
    })

    this.accountService.all().subscribe(users => {
      this.users = users;
    })
  }

  addNew() {
    this.users.push({
      id: 0,
      name: "New User",
      permissions: 0,
    })
  }

  handleDeleteUser(id: number) {
    this.users = this.users.filter(user => user.id !== id)
  }

  handleUpdateId(event: {old: number, new: number}) {
    this.users = this.users.map(user => {
      if (user.id === event.old) {
        user.id = event.new
      }
      return user;
    })
  }

  emptyUserPresent() {
    return this.users.find(user => user.id === 0) !== undefined;
  }

  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
}
