import {Component, Input} from '@angular/core';
import {FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";
import {NgIcon} from "@ng-icons/core";

@Component({
    selector: 'app-form-select',
    imports: [
        FormsModule,
        ReactiveFormsModule
    ],
    templateUrl: './form-select.component.html',
    styleUrl: './form-select.component.css'
})
export class FormSelectComponent {

  @Input({required: true}) formGroup!: FormGroup;
  @Input({required: true}) controlName!: string;
  @Input({required: true}) label!: string;
  @Input({required: true}) options!: string[];
  @Input() values: string[] | undefined;
  @Input() formGroupName: string | undefined;

  constructor() {

  }

}
