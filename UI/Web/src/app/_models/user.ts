export interface UserDto {
  id: number;
  name: string;
  email: string;
  permissions: number;
  canDelete: boolean;
}

export interface User {
  id: number;
  name: string;
  email: string;
  oidcToken?: string;
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

export function hasPermission(user: User | UserDto, perm: Perm): boolean {
  return (user.permissions & perm) === perm
}

export function roles(user: User | UserDto): Perm[] {
  return AllPerms.filter(val => hasPermission(user, val))
}

export const AllPerms = Object.values(Perm)
  .filter(value => typeof value === 'number')
  .filter(val => val !== 0) as number[];

