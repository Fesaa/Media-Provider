import {ChangeDetectorRef, Component, Input} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {FormGroup} from "@angular/forms";
import {Page} from "../../../_models/page";
import {DownloadService} from "../../../_services/download.service";
import {DownloadRequest} from "../../../_models/search";
import {bounceIn200ms} from "../../../_animations/bounce-in";
import {NgIcon} from "@ng-icons/core";
import {dropAnimation} from "../../../_animations/drop-animation";
import {ToastrService} from "ngx-toastr";
import {ImageService} from "../../../_services/image.service";

@Component({
    selector: 'app-search-result',
    imports: [
        NgIcon
    ],
    templateUrl: './search-result.component.html',
    styleUrl: './search-result.component.css',
    animations: [bounceIn200ms, dropAnimation]
})
export class SearchResultComponent {

  @Input({required: true}) page!: Page;
  @Input({required: true}) form!: FormGroup;
  @Input({required: true}) searchResult!: SearchInfo;

  showExtra: boolean = false;

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


  constructor(private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
              private toastR: ToastrService,
              private imageService: ImageService,
  ) {
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

    const customDir = this.form.value["customDir"];
    const dir = customDir ? customDir : this.form.value["dir"];

    const req: DownloadRequest = {
      provider: this.searchResult.Provider,
      title: this.searchResult.Name,
      id: this.searchResult.InfoHash,
      dir: dir,
    }

    this.downloadService.download(req).subscribe({
      complete: () => {
        this.toastR.success(`Downloaded started for ${this.searchResult.Name}`, "Success")
      },
      error: (err) => {
        this.toastR.error(`Download failed ${err.error.error}`, "Error")
      }
  })
  }

  getColour(idx: number): string {
    return this.colours[idx % this.colours.length];
  }

  toggleExtra() {
    this.showExtra = !this.showExtra;
    if (this.imageSource == null) {
      this.loadImage();
    }

    this.cdRef.detectChanges();
  }

}
