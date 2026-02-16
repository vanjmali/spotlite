import { Component, computed, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-pagination',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './pagination.html',
  styleUrl: './pagination.scss',
})
export class PaginationComponent {
  readonly pageSg = input<number>(1, { alias: 'page' });
  readonly pageSizeSg = input<number>(10, { alias: 'pageSize' });
  readonly totalSg = input<number>(0, { alias: 'total' });
  readonly pageSizeOptionsSg = input<number[]>([10, 20, 50], { alias: 'pageSizeOptions' });
  readonly disabledSg = input<boolean>(false, { alias: 'disabled' });

  readonly pageChange = output<number>();
  readonly pageSizeChange = output<number>();

  readonly totalPagesSg = computed(() => {
    const total = this.totalSg();
    const size = this.pageSizeSg();
    if (size <= 0) return 1;
    return Math.max(1, Math.ceil(total / size));
  });

  readonly rangeStartSg = computed(() => {
    const total = this.totalSg();
    if (total === 0) return 0;
    return (this.pageSg() - 1) * this.pageSizeSg() + 1;
  });

  readonly rangeEndSg = computed(() => {
    const total = this.totalSg();
    if (total === 0) return 0;
    return Math.min(this.pageSg() * this.pageSizeSg(), total);
  });

  readonly pagesSg = computed(() => {
    const totalPages = this.totalPagesSg();
    const current = this.pageSg();
    const pages: number[] = [];

    if (totalPages <= 5) {
      for (let i = 1; i <= totalPages; i += 1) pages.push(i);
      return pages;
    }

    const start = Math.max(1, current - 1);
    const end = Math.min(totalPages, current + 1);

    if (start > 1) pages.push(1);
    if (start > 2) pages.push(-1);

    for (let i = start; i <= end; i += 1) pages.push(i);

    if (end < totalPages - 1) pages.push(-1);
    if (end < totalPages) pages.push(totalPages);

    return pages;
  });

  goToPage(page: number): void {
    if (this.disabledSg()) return;
    const totalPages = this.totalPagesSg();
    const nextPage = Math.min(Math.max(1, page), totalPages);
    if (nextPage === this.pageSg()) return;
    this.pageChange.emit(nextPage);
  }

  goPrev(): void {
    this.goToPage(this.pageSg() - 1);
  }

  goNext(): void {
    this.goToPage(this.pageSg() + 1);
  }

  onPageSizeChange(event: Event): void {
    const value = (event.target as HTMLSelectElement | null)?.value ?? '';
    const nextSize = Number(value);
    if (!Number.isFinite(nextSize) || nextSize <= 0) return;
    if (nextSize === this.pageSizeSg()) return;
    this.pageSizeChange.emit(nextSize);
  }

  isEllipsis(page: number): boolean {
    return page === -1;
  }
}
