import {Pipe, PipeTransform} from '@angular/core';
import {Perm} from "../_models/user";

@Pipe({
  name: 'permNamePipe'
})
export class PermNamePipe implements PipeTransform {

  transform(value: Perm): string {
    switch (value) {
      case Perm.All:
        return 'All';
      case Perm.DeletePage:
        return 'Delete Page';
      case Perm.DeleteUser:
        return 'Delete User';
      case Perm.WriteConfig:
        return 'Write Config';
      case Perm.WritePage:
        return 'Write Page';
      case Perm.WriteUser:
        return 'Write User';
      default:
        return 'Unknown';
    }
  }

}
