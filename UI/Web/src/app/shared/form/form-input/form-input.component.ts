import {Component, Input} from '@angular/core';
import {FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";

@Component({
    selector: 'app-form-input',
    imports: [
        FormsModule,
        ReactiveFormsModule
    ],
    templateUrl: './form-input.component.html',
    styleUrl: './form-input.component.css'
})
export class FormInputComponent {

  @Input({required: true}) formGroup!: FormGroup;
  @Input({required: true}) controlName!: string;
  @Input({required: true}) label!: string;
  @Input() type: string = "text";
  @Input() formGroupName: string | undefined;



}
