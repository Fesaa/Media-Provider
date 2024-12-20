import { Pipe, PipeTransform } from '@angular/core';
import {RefreshFrequency} from "../_models/subscription";

@Pipe({
  name: 'refreshFrequency',
  standalone: true
})
export class RefreshFrequencyPipe implements PipeTransform {

  transform(value: RefreshFrequency): string {
    switch (value) {
      case RefreshFrequency.OneHour:
        return '1 Hour';
      case RefreshFrequency.HalfDay:
        return '12 Hours';
      case RefreshFrequency.FullDay:
        return '1 Day';
      case RefreshFrequency.Week:
        return '1 Week';
      default:
        return 'Unknown';
    }
  }

}
