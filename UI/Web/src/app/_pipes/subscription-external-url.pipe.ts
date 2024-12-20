import {Pipe, PipeTransform} from '@angular/core';
import {Provider} from "../_models/page";

@Pipe({
  name: 'subscriptionExternalUrl',
  standalone: true
})
export class SubscriptionExternalUrlPipe implements PipeTransform {

  transform(contentId: string, provider: Provider): string {
    switch (provider) {
      case Provider.MANGADEX:
        return "https://mangadex.org/title/" + contentId;
      case Provider.WEBTOON:
        return "https://www.webtoons.com/episodeList?titleNo=" + contentId;
      default:
        throw new Error(`Unsupported provider: ${provider}`);
    }
  }

}
