import {Component, HostListener, Input, OnInit} from '@angular/core';
import {ReplaySubject} from "rxjs";
import {NgClass} from "@angular/common";
import {Dialog} from "primeng/dialog";
import {Button} from "primeng/button";

@Component({
    selector: 'app-dialog',
  imports: [
    NgClass,
    Dialog,
    Button
  ],
    templateUrl: './dialog.component.html',
    styleUrl: './dialog.component.css'
})
export class DialogComponent implements OnInit {

  @Input() isMobile = false;
  @Input() text: string = '';
  @Input() header: string = '';

  visible: boolean = true;
  private result = new ReplaySubject<boolean>(1)



  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  ngOnInit(): void {
    this.isMobile = window.innerWidth < 768;
  }

  public getResult() {
    return this.result.asObservable();
  }

  closeDialog() {
    this.result.next(false);
    this.result.complete();
  }

  confirm() {
    this.result.next(true);
    this.result.complete();
  }

}
