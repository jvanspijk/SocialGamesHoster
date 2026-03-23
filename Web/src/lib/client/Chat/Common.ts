export type Message = {
	readonly id: string;
	readonly content: string;
	readonly sentAt: string;
	readonly senderId: number | null;
	readonly senderName: string | null;
};
