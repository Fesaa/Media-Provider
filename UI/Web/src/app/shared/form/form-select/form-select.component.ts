import {Component, computed, ContentChild, EventEmitter, input, Input, model, Output, TemplateRef} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";
import {translate} from "@jsverse/transloco";
import {Tooltip} from "primeng/tooltip";
import {FormItemComponent} from "../form-item/form-item.component";
import {NgTemplateOutlet} from "@angular/common";

@Component({
  selector: 'app-form-select',
  imports: [
    FormsModule,
    ReactiveFormsModule,
    Tooltip,
    NgTemplateOutlet
  ],
  templateUrl: './form-select.component.html',
  styleUrl: './form-select.component.scss'
})
export class FormSelectComponent extends FormItemComponent{

  @ContentChild('options') optionsRef!: TemplateRef<any>;

}
