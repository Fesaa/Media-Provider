
export enum Role {
  ManagePages          = "manage-pages",
  ManageUsers          = "manage-users",
  ManageServerConfigs  = "manage-server-configs",
  ManagePreferences    = "manage-preferences",
  ManageSubscriptions  = "manage-subscriptions",
  ViewAllDownloads     = "view-all-downloads",
}

export const AllRoles = [
  Role.ManagePages, Role.ManageUsers, Role.ManageServerConfigs, Role.ManagePreferences, Role.ManageSubscriptions,
  Role.ViewAllDownloads
]

export interface UserDto {
  id: number;
  name: string;
  email: string;
  roles: Role[];
  pages: number[];
  canDelete: boolean;
}

export interface User {
  id: number;
  name: string;
  email: string;
  oidcToken?: string;
  token: string;
  apiKey: string;
  roles: Role[];
}

export function hasRole(user: User, role: Role) {
  return user.roles.includes(role);
}

