import {AbstractControl, FormControl, ValidatorFn} from "@angular/forms";

export class IntegerFormControl extends FormControl {
  override setValue(value: any, options: any) {
    if (typeof value === 'string' && !isNaN(value as any)) {
      value = parseInt(value, 10);
    }
    super.setValue(value, options);
  }
}

export function BoundNumberValidator(min: number, max: number): ValidatorFn {
  return (control: AbstractControl): { [key: string]: any } | null => {
    if (control.value !== null && (isNaN(control.value) || control.value < min || control.value > max)) {
      return { 'boundNumber': { value: control.value } };
    }
    return null;
  };
}
