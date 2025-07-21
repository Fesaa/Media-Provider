import {Component, ContentChild, TemplateRef} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
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
