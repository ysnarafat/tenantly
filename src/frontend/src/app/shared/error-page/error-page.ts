import { Component, Input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

/**
 * Presentational error page shared by the 403 and 404 screens.
 *
 * The signature element is a door/unit nameplate bearing the status code — a nod
 * to the building-and-unit world Tenantly manages. Callers supply the copy and a
 * tone, and project their own action buttons into the default slot.
 */
@Component({
  selector: 'app-error-page',
  standalone: true,
  imports: [MatIconModule],
  templateUrl: './error-page.html',
  styleUrls: ['./error-page.scss'],
})
export class ErrorPage {
  /** HTTP status code shown on the nameplate, e.g. '403'. */
  @Input({ required: true }) code = '';
  /** Short status word beneath the code, e.g. 'Forbidden'. */
  @Input({ required: true }) status = '';
  /** Material icon on the badge, e.g. 'lock'. */
  @Input({ required: true }) icon = '';
  /** Headline explaining what happened. */
  @Input({ required: true }) title = '';
  /** One line of guidance on what to do next. */
  @Input({ required: true }) message = '';
  /** Accent tone: 'warn' (amber, for restricted) or 'primary' (blue). */
  @Input() tone: 'primary' | 'warn' = 'primary';
}
