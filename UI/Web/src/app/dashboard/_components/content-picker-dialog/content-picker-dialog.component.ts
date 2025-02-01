import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {InfoStat} from "../../../_models/stats";
import {ContentService} from "../../../_services/content.service";
import {ListContentData} from "../../../_models/messages";
import {TreeNode} from "primeng/api";
import {Tree} from "primeng/tree";
import {Button} from "primeng/button";
import {MessageService} from "../../../_services/message.service";
import {Dialog} from "primeng/dialog";
import {Skeleton} from "primeng/skeleton";
import {NgForOf} from "@angular/common";

@Component({
  selector: 'app-content-picker-dialog',
  imports: [
    Tree,
    Button,
    Dialog,
    Skeleton,
    NgForOf
  ],
  templateUrl: './content-picker-dialog.component.html',
  styleUrl: './content-picker-dialog.component.css'
})
export class ContentPickerDialogComponent {

  @Input({required: true}) visible!: boolean;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();
  @Input({required: true}) info!: InfoStat;

  content: ListContentData[] = [];
  selection: ListContentData[] = [];
  loading: boolean = true;


  constructor(
    private contentService: ContentService,
    private msgService: MessageService,
  ) {
  }

  loadContent() {
    this.loading = true;
    this.contentService.listContent(this.info.provider, this.info.id).subscribe(contents => {
      this.content = contents;
      this.selection = this.flatten(contents);
      this.loading = false;
    })
  }

  unselectAll() {
    this.selection = [];
  }

  selectAll() {
    this.selection = this.flatten(this.content)
  }

  reverse() {
    this.content = this.content.reverse();
  }

  close() {
    this.visibleChange.emit(false);
  }

  submit() {
    const ids = this.getAllSubContentIds(this.selection);

    if (ids.length == 0) {
      this.msgService.warning("Not saving", "Remove the content if you wish nothing to be downloaded")
      return;
    }

    this.contentService.setFilter(this.info.provider, this.info.id, ids).subscribe({
      next: () => {
        this.msgService.success("Succes", `Set filter to ${ids.length} items`);
      },
      error: (err) => {
        this.msgService.error("Error", `Failed:\n ${err.error.message}`);
      }
    }).add(() => (
      this.close()
    ))
  }

  private flatten(list: ListContentData[]): ListContentData[] {
    const result: ListContentData[] = [];

    function isFullySelected(data: ListContentData & TreeNode): boolean {
      if (data.subContentId && !data.selected) {
        return false;
      }

      if (data.subContentId && data.selected) {
        return true;
      }

      if (!data.children) {
        // Empty directory?
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

      if (atLeastOne && !allSelected) {
        data.partialSelected = true;
      }

      return allSelected || atLeastOne;
    }

    function traverse(items: ListContentData[]) {
      for (const item of items) {

        if (isFullySelected(item)) {
          result.push(item);
        }

        if (item.children && item.children.length > 0) {
          traverse(item.children);
        }
      }
    }

    traverse(list);
    return result;
  }


  private getAllSubContentIds(list: ListContentData[]): string[] {
    const result: string[] = [];

    function traverse(items: ListContentData[]) {
      for (const item of items) {
        if (item.subContentId !== undefined && !result.includes(item.subContentId)) {
          result.push(item.subContentId);
        }
        if (item.children && item.children.length > 0) {
          traverse(item.children);
        }
      }
    }

    traverse(list);
    return result;
  }


}
