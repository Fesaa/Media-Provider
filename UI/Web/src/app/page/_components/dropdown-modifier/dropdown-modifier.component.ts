import {Component, Input} from '@angular/core';
import {KeyValuePipe} from "@angular/common";
import {ReactiveFormsModule} from "@angular/forms";
import {Modifier} from "../../../_models/page";

@Component({
    selector: 'app-dropdown-modifier',
    imports: [
        KeyValuePipe,
        ReactiveFormsModule
    ],
    templateUrl: './dropdown-modifier.component.html',
    styleUrl: './dropdown-modifier.component.css'
})
export class DropdownModifierComponent {

  @Input({required: true}) key!: string;
  @Input({required: true}) modifier!: Modifier;

}
