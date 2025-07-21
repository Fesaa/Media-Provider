import {
  Component,
  OnInit,
  ChangeDetectionStrategy,
  signal,
  computed,
  inject,
  input
} from '@angular/core';
import { InfoStat } from '../../../_models/stats';
import { ContentService } from '../../../_services/content.service';
import { ListContentData } from '../../../_models/messages';
import { ToastService } from '../../../_services/toast.service';
import { TranslocoDirective } from '@jsverse/transloco';
import {NgbActiveModal} from "@ng-bootstrap/ng-bootstrap";

@Component({
  selector: 'app-content-picker-dialog',
  standalone: true,
  imports: [TranslocoDirective],
  templateUrl: './content-picker-dialog.component.html',
  styleUrls: ['./content-picker-dialog.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ContentPickerDialogComponent implements OnInit {

  private readonly contentService = inject(ContentService);
  private readonly toastService = inject(ToastService);
  private readonly modal = inject(NgbActiveModal);

  info = input.required<InfoStat>();

  content = signal<ListContentData[]>([]);
  selection = signal<ListContentData[]>([]);
  loading = signal(true);

  // Derived state for selection count (example of computed usage)
  selectionCount = computed(() => this.selection().length);

  ngOnInit(): void {
    this.loading.set(true);
    this.contentService.listContent(this.info().provider, this.info().id).subscribe({
      next: contents => {
        this.content.set(contents);
        this.selection.set(this.flatten(contents));
        this.loading.set(false);
      },
      error: err => {
        this.toastService.genericError(err?.error?.message ?? 'Unknown error');
        this.loading.set(false);
      }
    });
  }

  unselectAll(): void {
    this.selection.set([]);
  }

  selectAll(): void {
    this.selection.set(this.flatten(this.content()));
  }

  reverse(): void {
    this.content.update(c => [...c].reverse());
  }

  close(): void {
    this.modal.close();
  }

  submit(): void {
    const ids = this.getAllSubContentIds(this.selection());

    if (ids.length === 0) {
      this.toastService.warningLoco('dashboard.content-picker.toasts.no-changes');
      return;
    }

    this.contentService.setFilter(this.info().provider, this.info().id, ids).subscribe({
      next: () => {
        this.toastService.successLoco(
          'dashboard.content-picker.toasts.success',
          {},
          { amount: ids.length, title: this.info.name }
        );
      },
      error: err => {
        this.toastService.genericError(err?.error?.message ?? 'Unknown error');
      },
    }).add(() => {
      this.close();
    });
  }

  private flatten(list: ListContentData[]): ListContentData[] {
    const result: ListContentData[] = [];

    function isFullySelected(data: ListContentData): boolean {
      if (data.subContentId && !data.selected) {
        return false;
      }
      if (data.subContentId && data.selected) {
        return true;
      }
      if (!data.children?.length) {
        return false;
      }
      let allSelected = true;
      let atLeastOne = false;
      for (const child of data.children) {
        if (isFullySelected(child)) {
          atLeastOne = true;
        } else {
          allSelected = false;
        }
      }
      data.partialSelected = atLeastOne && !allSelected;
      return allSelected || atLeastOne;
    }

    function traverse(items: ListContentData[]): void {
      for (const item of items) {
        if (isFullySelected(item)) {
          result.push(item);
        }
        if (item.children?.length) {
          traverse(item.children);
        }
      }
    }

    traverse(list);
    return result;
  }

  private getAllSubContentIds(list: ListContentData[]): string[] {
    const result: string[] = [];

    function traverse(items: ListContentData[]): void {
      for (const item of items) {
        if (item.subContentId && !result.includes(item.subContentId)) {
          result.push(item.subContentId);
        }
        if (item.children?.length) {
          traverse(item.children);
        }
      }
    }

    traverse(list);
    return result;
  }
}
