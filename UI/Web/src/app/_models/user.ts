
export interface User {
  token: string;
  apiKey: string;
  permissions: number;
}

export enum Perm {
  All = 0,

  WritePage = 1 << 0,
  DeletePage = 1 << 1,

  WriteUser = 1 << 2,
  DeleteUser = 1 << 3,

  WriteConfig = 1 << 4,
}

export function hasPermission(user: User, perm: Perm): boolean {
  return (user.permissions&perm)===perm
}
