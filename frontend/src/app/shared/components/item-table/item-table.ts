import { Component, input, ViewEncapsulation } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WidgetComponent } from '@app/shared/components/widget';
import { LoaderComponent } from '@app/shared/components/loader';

@Component({
  selector: 'app-item-table',
  standalone: true,
  imports: [CommonModule, WidgetComponent, LoaderComponent],
  templateUrl: './item-table.html',
  styleUrl: './item-table.scss',
  encapsulation: ViewEncapsulation.None,
})
export class ItemTableComponent {
  readonly titleSg = input<string>('');
  readonly isLoadingSg = input<boolean>(false, { alias: 'isLoading' });
  readonly isEmptySg = input<boolean>(false, { alias: 'isEmpty' });
  readonly emptyMessageSg = input<string>('No items found.', { alias: 'emptyMessage' });
  readonly showPaginationSg = input<boolean>(false, { alias: 'showPagination' });
}
