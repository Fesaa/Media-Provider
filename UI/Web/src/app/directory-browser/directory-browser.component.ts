import {ChangeDetectorRef, Component, Input, OnInit} from '@angular/core';
import {IoService} from "../_services/io.service";
import {ToastrService} from "ngx-toastr";
import {DirEntry} from "../_models/io";
import {FormsModule} from "@angular/forms";
import {Stack} from "../shared/data-structures/stack";
import {NgIcon} from "@ng-icons/core";
import {Clipboard} from "@angular/cdk/clipboard";
import {dropAnimation} from "../_animations/drop-animation";

@Component({
    selector: 'app-directory-browser',
    imports: [
        FormsModule,
        NgIcon
    ],
    templateUrl: './directory-browser.component.html',
    styleUrl: './directory-browser.component.css',
    animations: [dropAnimation]
})
export class DirectoryBrowserComponent implements OnInit{

  @Input({required: true}) root!: string;
  @Input() showFiles: boolean = false;
  @Input() filter: boolean = false;
  @Input() copy: boolean = true;
  @Input() create: boolean = false;

  currentRoot = '';
  entries: DirEntry[] = [];
  routeStack: Stack<string> = new Stack<string>();

  query: string = '';
  newDirName: string = '';

  constructor(private ioService: IoService,
              private toastR: ToastrService,
              private clipboard: Clipboard,
  ) {}

  ngOnInit(): void {
    this.currentRoot = this.root;
    this.routeStack.push(this.root);
    this.loadChildren(this.root);
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
    this.loadChildren( this.routeStack.items.join('/'));
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

  private loadChildren(dir: string) {
    this.ioService.ls(dir, this.showFiles).subscribe({
      next: (entries) => {
        this.entries = entries || [];
      },
      error: (err) => {
        this.routeStack.pop();
        console.error(err);
      }
    })
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
        this.toastR.error(`Failed to create directory ${this.newDirName}. \n ${err.error.error}`, 'Error');
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



}
