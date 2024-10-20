import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';
import {NgIcon} from "@ng-icons/core";

@Component({
  selector: 'app-paginator',
  templateUrl: './paginator.component.html',
  imports: [
    NgIcon,
  ],
  standalone: true
})
export class PaginatorComponent implements OnInit {
  @Input() totalPages: number = 1;
  @Input() currentPage: number = 1;
  @Output() pageChange = new EventEmitter<number>();

  pages: number[] = [];

  ngOnInit() {
    this.pages = Array.from({ length: this.totalPages }, (_, i) => i + 1);
  }

  goToPage(page: number) {
    if (page >= 1 && page <= this.totalPages) {
      this.currentPage = page;
      this.pageChange.emit(this.currentPage);
    }
  }

  previousPage() {
    if (this.currentPage > 1) {
      this.currentPage--;
      this.pageChange.emit(this.currentPage);
    }
  }

  nextPage() {
    if (this.currentPage < this.totalPages) {
      this.currentPage++;
      this.pageChange.emit(this.currentPage);
    }
  }
}
