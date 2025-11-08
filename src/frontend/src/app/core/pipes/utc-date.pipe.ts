import { Pipe, PipeTransform, inject } from '@angular/core';
import { DateUtilsService } from '../services/date-utils.service';

@Pipe({
  name: 'utcDate',
  standalone: true
})
export class UtcDatePipe implements PipeTransform {
  private dateUtils = inject(DateUtilsService);

  transform(
    value: string | null | undefined, 
    format: 'short' | 'medium' | 'long' | 'full' | 'date' | 'time' | 'relative' = 'medium'
  ): string {
    if (!value) return '';

    switch (format) {
      case 'date':
        return this.dateUtils.formatDateOnly(value);
      case 'time':
        return this.dateUtils.formatTimeOnly(value);
      case 'relative':
        return this.dateUtils.getRelativeTime(value);
      default:
        return this.dateUtils.formatToLocal(value, format);
    }
  }
}
