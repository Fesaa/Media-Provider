import {Component, HostListener, Input, OnInit} from '@angular/core';
import {ReplaySubject} from "rxjs";
import {NgIcon} from "@ng-icons/core";
import {DirEntry} from "../../../_models/io";
import {Stack} from "../../data-structures/stack";
import {IoService} from "../../../_services/io.service";
import {ToastrService} from "ngx-toastr";
import {Clipboard} from "@angular/cdk/clipboard";
import {FormsModule} from "@angular/forms";
import {Dialog} from "primeng/dialog";
import {Button} from "primeng/button";

@Component({
    selector: 'app-directory-selector',
  imports: [
    NgIcon,
    FormsModule,
    Dialog,
    Button
  ],
    templateUrl: './directory-selector.component.html',
    styleUrl: './directory-selector.component.css'
})
export class DirectorySelectorComponent implements OnInit {

  @Input() isMobile = false;

  @Input({required: true}) root!: string;
  @Input() showFiles: boolean = false;
  @Input() filter: boolean = false;
  @Input() copy: boolean = true;
  @Input() create: boolean = false;
  @Input() customWidth: string = '50vw';

  currentRoot = '';
  entries: DirEntry[] = [];
  routeStack: Stack<string> = new Stack<string>();

  query: string = '';
  newDirName: string = '';
  visible: boolean = true;
  private result = new ReplaySubject<string | undefined>(1)

  constructor(private ioService: IoService,
              private toastR: ToastrService,
              private clipboard: Clipboard,
  ) {
  }

  ngOnInit(): void {
    this.currentRoot = this.root;
    this.routeStack.push(this.root);
    this.loadChildren(this.root);
    this.isMobile = window.innerWidth < 768;
  }

  getEntries() {
    return this.entries.filter(entry => this.normalize(entry.name).includes(this.query));
  }

  selectNode(entry: DirEntry) {
    if (!entry.dir) {
      return;
    }

    this.query = '';
    this.currentRoot = entry.name;
    this.routeStack.push(entry.name);
    this.loadChildren(this.routeStack.items.join('/'));
  }

  goBack() {
    if (this.routeStack.isEmpty()) {
      return;
    }

    this.routeStack.pop();
    const nextRoot = this.routeStack.peek();
    if (nextRoot) {
      this.currentRoot = nextRoot;
    }
    this.loadChildren(this.routeStack.items.join('/'));
  }

  normalize(str: string): string {
    return str.toLowerCase();
  }

  onFilterChange(event: Event) {
    const inputElement = event.target as HTMLInputElement;
    this.query = this.normalize(inputElement.value);
  }

  onNewDirNameChange(event: Event) {
    const inputElement = event.target as HTMLInputElement;
    this.newDirName = inputElement.value;
  }

  createNew() {
    this.ioService.create(this.routeStack.items.join('/'), this.newDirName).subscribe({
      next: () => {
        this.toastR.success(`Directory ${this.newDirName} created successfully`, 'Success');
        this.newDirName = '';
        this.loadChildren(this.routeStack.items.join('/'));
      },
      error: (err) => {
        this.toastR.error(`Failed to create directory ${this.newDirName}\n${err.error.message}`, 'Error');
        console.error(err);
      }
    });
  }

  copyPath(entry: DirEntry) {
    let path = this.routeStack.items.join('/');
    if (entry.dir) {
      path += '/' + entry.name;
    }
    this.clipboard.copy(path);
  }

  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  public getResult() {
    return this.result.asObservable();
  }

  closeDialog() {
    this.result.next(undefined);
    this.result.complete();
    this.visible = false;
  }

  confirm() {
    let path = this.routeStack.items.join('/');
    if (path.startsWith('/')) {
      path = path.substring(1);
    }
    this.result.next(path);
    this.result.complete();
    this.visible = false;
  }

  private loadChildren(dir: string) {
    this.ioService.ls(dir, this.showFiles).subscribe({
      next: (entries) => {
        this.entries = entries || [];
      },
      error: (err) => {
        this.routeStack.pop();
        console.error(err);
        this.toastR.error(err.error.message, "Failed to load children")
      }
    })
  }

}
