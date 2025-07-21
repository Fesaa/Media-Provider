import {ChangeDetectorRef, Component, inject, input, Input, OnInit, signal} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {DownloadMetadata, Page, Provider} from "../../../_models/page";
import {bounceIn200ms} from "../../../_animations/bounce-in";
import {dropAnimation} from "../../../_animations/drop-animation";
import {ImageService} from "../../../_services/image.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {NgStyle} from "@angular/common";
import {NgbTooltip} from "@ng-bootstrap/ng-bootstrap";
import {ModalService} from "../../../_services/modal.service";
import {DownloadModalComponent} from "../download-modal/download-modal.component";
import {DefaultModalOptions} from "../../../_models/default-modal-options";

@Component({
  selector: 'app-search-result',
  imports: [
    TranslocoDirective,
    NgStyle,
    NgbTooltip,
  ],
  templateUrl: './search-result.component.html',
  styleUrl: './search-result.component.scss',
  animations: [bounceIn200ms, dropAnimation]
})
export class SearchResultComponent implements OnInit{

  private readonly imageService = inject(ImageService);
  private readonly modalService = inject(ModalService);

  page = input.required<Page>();
  searchResult = input.required<SearchInfo>();
  dir = input.required<string>();
  providers = input.required<Provider[]>();
  metadata = input.required<DownloadMetadata>();

  imageSource = signal<string | null>(null);


  ngOnInit(): void {
    this.loadImage();
  }

  addAsSub() {

  }

  download() {
    const metadata = this.metadata();
    if (!metadata) return

    const [_, component] = this.modalService.open(DownloadModalComponent, DefaultModalOptions);
    component.metadata.set(metadata);
    component.page.set(this.page());
    component.info.set(this.searchResult());
  }

  loadImage() {
    if (this.searchResult().ImageUrl === "") {
      return;
    }

    if (this.searchResult().ImageUrl.startsWith("proxy")) {
      this.imageService.getImage(this.searchResult().ImageUrl).subscribe(src => {
        this.imageSource.set(src);
      })
    } else {
      this.imageSource.set(this.searchResult().ImageUrl);
    }
  }

}
