import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {of, Subject} from "rxjs";
import {ToastrService} from "ngx-toastr";
import {AccountService} from "./account.service";

@Injectable({
  providedIn: 'root'
})
export class ImageService {

  baseUrl = environment.apiUrl;
  apiKey: string | null = null;

  constructor(private httpClient: HttpClient, private toastR: ToastrService, private accountService: AccountService) {
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.apiKey = user.apiKey;
      }
    })
  }

  getImage(imageUrl: string) {
    if (this.apiKey) {
      return of(this.baseUrl + imageUrl + `?api-key=${this.apiKey}`);
    }

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
