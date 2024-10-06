import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";

@Injectable({
  providedIn: 'root'
})
export class ImageService {

  baseUrl = environment.apiUrl;

  constructor(private httpClient: HttpClient) { }

  getImage(imageUrl: string) {
    return this.httpClient.get(this.baseUrl + imageUrl, { responseType: 'blob' });
  }
}
