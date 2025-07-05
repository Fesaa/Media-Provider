import {Component, computed, input, Input} from '@angular/core';
import {FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";
import {translate} from "@jsverse/transloco";
import {Tooltip} from "primeng/tooltip";

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
export class FormInputComponent {

  @Input({required: true}) formGroup!: FormGroup;
  @Input({required: true}) controlName!: string;
  @Input() label: string = '';
  @Input() type: string = "text";
  @Input() formGroupName: string | undefined = undefined;

  translationKey = input('');

  toolTip = computed(() => {
    const key = this.translationKey();
    if (!key || key === '' ) {
      return undefined;
    }

    const toolTipKey = key + '.tooltip';
    const translation = translate(toolTipKey);
    if (translation && translation !== toolTipKey) {
      return translation;
    }

    return undefined;
  })

  protected readonly translate = translate;
}
