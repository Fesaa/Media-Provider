import { Injectable } from '@angular/core';
import {ReplaySubject} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class NavService {

  private showNavSource = new ReplaySubject<Boolean>(1);
  public showNav$ = this.showNavSource.asObservable();

  constructor() {
    this.showNavSource.next(false);
  }

  setNavVisibility(show: Boolean) {
    this.showNavSource.next(show);
  }
}
