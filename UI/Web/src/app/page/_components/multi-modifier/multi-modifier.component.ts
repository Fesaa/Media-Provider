import {ChangeDetectorRef, Component, Input} from '@angular/core';
import {Modifier} from "../../../_models/page";
import {FormGroup} from "@angular/forms";
import {KeyValuePipe} from "@angular/common";
import {NgIcon} from "@ng-icons/core";
import {animate, query, sequence, stagger, style, transition, trigger} from "@angular/animations";
import {dropAnimation} from "../../../_animations/drop-animation";

@Component({
  selector: 'app-multi-modifier',
  standalone: true,
  imports: [
    KeyValuePipe,
    NgIcon
  ],
  templateUrl: './multi-modifier.component.html',
  styleUrl: './multi-modifier.component.css',
  animations: [dropAnimation]
})
export class MultiModifierComponent {

  @Input({required: true}) form!: FormGroup;
  @Input({required: true}) key!: string;
  @Input({required: true}) modifier!: Modifier;

  isDropdownOpen: boolean = false;
  query = '';

  constructor(private cdRef: ChangeDetectorRef) {
  }



  toggleDropdown() {
    this.isDropdownOpen = !this.isDropdownOpen;
  }

  normalize(str: string): string {
    return str.toLowerCase();
  }

  onFilterChange(event: Event) {
    const inputElement = event.target as HTMLInputElement;
    this.query = this.normalize(inputElement.value);
  }

  onCheckboxChange(value: string) {
    const formArray = this.form.controls[this.key];
    if (formArray.value.includes(value)) {
      formArray.patchValue(formArray.value.filter((v: string) => v !== value));
    } else {
      formArray.patchValue([...formArray.value, value]);
    }
  }

  isChecked(value: string): boolean {
    return this.form.controls[this.key].value.includes(value);
  }

  size(): number {
    const formArray = this.form.controls[this.key];
    return formArray.value.length;
  }

}
