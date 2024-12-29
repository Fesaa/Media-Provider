import {Provider} from "./page";

export interface UserDto {
  id: number;
  name: string;
  permissions: number;
}

export interface User {
  ID: number;
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
  return (user.permissions&perm)===perm
}

export const permissionNames = Object.keys(Perm)
  .filter(key => isNaN(Number(key)))
  .filter(val => val !== "All") as string[];
export const permissionValues = Object.values(Perm)
  .filter(value => typeof value === 'number')
  .filter(val => val !== 0) as number[];

