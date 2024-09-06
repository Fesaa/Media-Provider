import { Injectable } from '@angular/core';
import {ReplaySubject} from "rxjs";
import {ActivatedRoute} from "@angular/router";

@Injectable({
  providedIn: 'root'
})
export class NavService {

  private showNavSource = new ReplaySubject<Boolean>(1);
  public showNav$ = this.showNavSource.asObservable();

  private pageIndexSource = new ReplaySubject<number>(1);
  public pageIndex$ = this.pageIndexSource.asObservable();

  constructor(private route: ActivatedRoute) {
    this.showNavSource.next(false);

    this.route.queryParams.subscribe(params => {
      const index = params['index'];
      if (index) {
        try {
          const i = parseInt(index);
          if (i >= 0) {
            this.pageIndexSource.next(i);
          }
        } catch (e) {
          console.error(e);
        }
      }
    })
  }

  setNavVisibility(show: Boolean) {
    this.showNavSource.next(show);
  }
}
