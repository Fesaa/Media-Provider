import {Component, input} from '@angular/core';
import { CommonModule } from '@angular/common';

type SpinnerSize = 'small' | 'medium' | 'large';
type SpinnerColour = 'primary' | 'secondary' | 'white';

@Component({
  selector: 'app-loading-spinner',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './loading-spinner.component.html',
  styleUrls: ['./loading-spinner.component.scss']
})
export class LoadingSpinnerComponent {
  size  = input<SpinnerSize>('medium');
  colour = input<SpinnerColour>('primary');
  centered = input(false);
  message = input('');
}
