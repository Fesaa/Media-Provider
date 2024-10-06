import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Subject} from "rxjs";
import {ToastrService} from "ngx-toastr";

@Injectable({
  providedIn: 'root'
})
export class ImageService {

  baseUrl = environment.apiUrl;

  constructor(private httpClient: HttpClient, private toastR: ToastrService) { }

  getImage(imageUrl: string) {
    const imageSrc = new Subject<string>();
    this.httpClient.get(this.baseUrl + imageUrl, { responseType: 'blob' }).subscribe({
      next: blob => {
        const reader = new FileReader();
        reader.onloadend = () => {
          imageSrc.next(reader.result as string);
        }
        reader.readAsDataURL(blob);
      },
      error: err => {
        console.error(err);
        this.toastR.error("Unable to download image " + imageUrl, "Error");
      }
    })
    return imageSrc.asObservable();
  }
}
