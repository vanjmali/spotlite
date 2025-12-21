import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MessageType } from '@app/shared/enums';

@Component({
  selector: 'app-message',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './message.html',
  styleUrls: ['./message.scss'],
})
export class MessageComponent {
  // Expose enum for template usage
  public MessageType = MessageType;
  /**
   * The type of message to display
   * @default 'info'
   */
  public readonly typeSg = input<MessageType | string>(MessageType.Info, { alias: 'type' });

  /**
   * The message text content
   */
  public readonly messageSg = input.required<string>({ alias: 'message' });
}
