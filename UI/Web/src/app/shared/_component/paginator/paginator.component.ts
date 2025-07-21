import {Component, computed, ContentChild, input, signal, TemplateRef} from '@angular/core';
import {NgTemplateOutlet} from "@angular/common";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-paginator',
  imports: [
    NgTemplateOutlet,
    TranslocoDirective
  ],
  templateUrl: './paginator.component.html',
  styleUrl: './paginator.component.scss'
})
export class PaginatorComponent<T> {

  @ContentChild("items") itemsTemplate!: TemplateRef<any>;

  items = input.required<T[]>();
  pageSize = input(10);

  currentPage = signal(1);
  totalPages = computed(() => Math.floor(this.items().length / this.pageSize())+1);
  visibleItems = computed(() => {
    const page = this.currentPage();
    const pageSize = this.pageSize();
    const items = this.items();

    return items.slice(page * pageSize, (page+1) * pageSize);
  })

  range = (n: number) => Array.from({ length: n}, (_, i) => i);

  goToPage(page: number): void {
    if (page >= 1 && page <= this.totalPages()) {
      this.currentPage.set(page);
    }
  }

  nextPage(): void {
    this.goToPage(this.currentPage() + 1);
  }

  prevPage(): void {
    this.goToPage(this.currentPage() - 1);
  }


}
