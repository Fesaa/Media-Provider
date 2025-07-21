import {ChangeDetectorRef, Component, Input, OnInit} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {DownloadMetadata, Page, Provider} from "../../../_models/page";
import {bounceIn200ms} from "../../../_animations/bounce-in";
import {dropAnimation} from "../../../_animations/drop-animation";
import {ImageService} from "../../../_services/image.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {NgStyle} from "@angular/common";
import {NgbTooltip} from "@ng-bootstrap/ng-bootstrap";

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

  @Input({required: true}) page!: Page;
  @Input({required: true}) searchResult!: SearchInfo;
  @Input({required: true}) dir!: string;
  @Input({required: true}) providers!: Provider[];
  @Input({required: true}) metadata!: DownloadMetadata | undefined;

  showExtra: boolean = false;
  showDownloadDialog: boolean = false;
  showSubscriptionDialog: boolean = false;

  colours = [
    "bg-blue-200 dark:bg-blue-800",
    "bg-green-200 dark:bg-green-800",
    "bg-yellow-200 dark:bg-yellow-700",
    "bg-red-200 dark:bg-red-800",
    "bg-purple-200 dark:bg-purple-800",
    "bg-pink-200 dark:bg-pink-800",
    "bg-indigo-200 dark:bg-indigo-800",
    "bg-gray-200 dark:bg-gray-700"
  ];

  imageSource: string | null = null;


  constructor(private cdRef: ChangeDetectorRef,
              private imageService: ImageService,
  ) {
  }

  ngOnInit(): void {
    this.loadImage();
    this.cdRef.markForCheck();
  }

  addAsSub() {
    this.showSubscriptionDialog = true;
  }

  loadImage() {
    if (this.searchResult.ImageUrl === "") {
      return;
    }

    if (this.searchResult.ImageUrl.startsWith("proxy")) {
      this.imageService.getImage(this.searchResult.ImageUrl).subscribe(src => {
        this.imageSource = src;
      })
    } else {
      this.imageSource = this.searchResult.ImageUrl;
    }
  }

  download() {
    this.showDownloadDialog = true;
  }

  getColour(idx: number): string {
    return this.colours[idx % this.colours.length];
  }

}
