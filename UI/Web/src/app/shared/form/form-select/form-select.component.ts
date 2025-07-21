import {Component, ContentChild, TemplateRef} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {FormItemComponent} from "../form-item/form-item.component";
import {KeyValuePipe, NgTemplateOutlet} from "@angular/common";
import {NgbTooltip} from "@ng-bootstrap/ng-bootstrap";
import {translate} from "@jsverse/transloco";

@Component({
  selector: 'app-form-select',
  imports: [
    FormsModule,
    ReactiveFormsModule,
    NgTemplateOutlet,
    NgbTooltip,
    KeyValuePipe
  ],
  templateUrl: './form-select.component.html',
  styleUrl: './form-select.component.scss'
})
export class FormSelectComponent extends FormItemComponent{

  @ContentChild('options') optionsRef!: TemplateRef<any>;

  protected readonly translate = translate;
}
