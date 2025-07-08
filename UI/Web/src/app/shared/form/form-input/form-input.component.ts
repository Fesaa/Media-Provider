import {Component, computed, input, Input, model} from '@angular/core';
import {AbstractControl, FormControl, FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";
import {translate} from "@jsverse/transloco";
import {Tooltip} from "primeng/tooltip";
import {FormItemComponent} from "../form-item/form-item.component";

@Component({
  selector: 'app-form-input',
  imports: [
    FormsModule,
    ReactiveFormsModule,
    Tooltip
  ],
  templateUrl: './form-input.component.html',
  styleUrl: './form-input.component.css'
})
export class FormInputComponent extends FormItemComponent {

  type = input('text');
}
