export class SortedList<T> {

  private list: T[] = []
  private readonly comparator: (a: T, b: T) => number = (a, b) => 0;

  constructor(compare: (a: T, b: T) => number) {
    this.comparator = compare;
  }

  add(t: T) {
    this.list.push(t);
    this.list.sort(this.comparator)
  }

  addAll(items: T[]) {
    this.list.push(...items);
    this.list.sort(this.comparator)
  }

  set(items: T[]) {
    this.list = items;
    this.list.sort(this.comparator)
  }

  setFunc(f: (t: T) => T) {
    this.list = this.list.map(f)
    this.list.sort(this.comparator)
  }

  includes(t: T): boolean {
    return this.list.includes(t)
  }

  remove(t: T) {
    const index = this.list.indexOf(t);
    if (index > -1) {
      this.list.splice(index, 1);
    }
  }

  removeFunc(f: (t: T) => boolean) {
    this.list = this.list.filter(t => !f(t));
  }

  get(idx: number) {
    return this.list[idx];
  }

  getFunc(f: (t: T) => boolean) {
    const item = this.list.find(x => f(x));
    return item
  }

  length(): number {
    return this.list.length;
  }

  items(): T[] {
    return this.list;
  }

  sort() {
    this.list.sort(this.comparator);
  }

}
