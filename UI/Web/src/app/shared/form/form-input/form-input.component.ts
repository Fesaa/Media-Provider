import {Component, input} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {TranslocoDirective} from "@jsverse/transloco";
import {Tooltip} from "primeng/tooltip";
import {FormItemComponent} from "../form-item/form-item.component";
import {KeyValuePipe} from "@angular/common";

@Component({
  selector: 'app-form-input',
  imports: [
    FormsModule,
    ReactiveFormsModule,
    Tooltip,
    KeyValuePipe,
    TranslocoDirective
  ],
  templateUrl: './form-input.component.html',
  styleUrl: './form-input.component.scss'
})
export class FormInputComponent extends FormItemComponent {

  type = input('text');
}
