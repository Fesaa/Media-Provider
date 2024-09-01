import {Provider} from "./page";


export type SearchInfo = {
  Name: string;
  Description: string;
  Date: string;
  Size: string;
  Seeders: number;
  Leechers: number;
  Downloads: number;
  Link: string;
  InfoHash: string;
  ImageUrl: string;
  RefUrl: string;
  Provider: Provider;
}
