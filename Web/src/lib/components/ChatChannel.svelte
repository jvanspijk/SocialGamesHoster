<script lang="ts">
	import TextInput from './TextInput.svelte';
	import MainButton from './MainButton.svelte';
    import { chatHub } from '$lib/client/Chat/ChatHub.svelte';
	import { GetMessage } from '$lib/client/Chat/GetMessage';
	import { onMount } from 'svelte';
	import { SendMessage } from '$lib/client/Chat/SendMessage';

	function generateID() {
		const timestamp = new Date().getTime();
		const randomNum = Math.floor(Math.random() * 10000);
		return `id-${timestamp}-${randomNum}`;
	}

	interface Message {
		id: string;
		text: string;
		sender: string;
		isMe: boolean;
		time: string;
	}

	let { 
        channelId, 
        channelName = "Global Chat",
		readerId,
        isOpen = false, 
        onClose
    } = $props<{
        channelId: number;
        channelName?: string;
		readerId: number;
        isOpen?: boolean;
        onClose: () => void;
    }>();

	let newMessage = $state('');
    let messages: Message[] = $state([]);
    let scrollContainer = $state<HTMLElement | null>(null);

	function formatTime(dateStr?: string) {
        const date = dateStr ? new Date(dateStr) : new Date();
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    onMount(() => {
        const unsubscribe = chatHub.onEvent('MessageSent', async (event) => {
			console.debug(`Message received: ${JSON.stringify(event)}`);
            if (event.channelId !== channelId) return;
			if(event.senderId === readerId) return;
			
            const response = await GetMessage(fetch, { id: event.messageId });
            
            if (response.ok) {
                const data = response.data;
                if (!messages.some(m => m.id === data.id)) {
                    messages.push({
                        id: data.id,
                        text: data.content,
                        sender: data.senderName,
                        isMe: data.senderId === readerId,
                        time: formatTime(data.sentAt)
                    });
                }
            }
        });

        return unsubscribe;
    });

    async function sendMessage() {
        if (!newMessage.trim()) return;

        const textToSend = newMessage;
        const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        messages.push({ id: generateID(), text: textToSend, sender: "You", isMe: true, time: timestamp });
        newMessage = '';

        try {
            await SendMessage(fetch, { channelId, playerId: readerId, message: textToSend });
        } catch (err) {
            console.error("Failed to send message:", err);
        }
    }

    $effect(() => {
        if (isOpen || messages.length) {
            if (scrollContainer) {
                scrollContainer.scrollTo({
                    top: scrollContainer.scrollHeight,
                    behavior: 'smooth'
                });
            }
        }
    });

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') onClose();
    }

	
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <div class="modal-overlay" onclick={(e) => e.target === e.currentTarget && onClose()} role="none">
        <div class="chat-modal">
            <header>
                <div class="channel-info">
                    <span class="status-dot"></span>
                    <h2>{channelName}</h2>
                </div>
                <button class="close-btn" onclick={onClose}>✕</button>
            </header>

            <div class="message-container" bind:this={scrollContainer}>
                {#each messages as msg (msg.id)}
                    <div class="message-wrapper {msg.isMe ? 'me' : 'them'}">
                        <div class="bubble">
                            {#if !msg.isMe}<span class="sender">{msg.sender}</span>{/if}
                            <p>{msg.text}</p>
                            <span class="timestamp">{msg.time}</span>
                        </div>
                    </div>
                {/each}
            </div>

            <footer>
                <div class="input-wrapper">
                    <TextInput 
                        bind:value={newMessage} 
                        placeholder="Type a message..." 
                        on:keydown={(e: KeyboardEvent) => e.key === 'Enter' && sendMessage()}
                    />
                </div>
                <div class="send-wrapper">
                    <MainButton label="Send" onActivate={sendMessage} />
                </div>
            </footer>
        </div>
    </div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.85);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 2000;
		backdrop-filter: blur(4px);
	}

	.chat-modal {
		width: 95vw;
		max-width: 600px;
		height: 85vh;
		background-color: #fff9e6; /* Parchment Color */
		border: 4px solid #5b4a3c;
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
	}

	header {
		background: #3e322b;
		padding: 15px 20px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 3px solid #5b4a3c;
		color: #f7e7c4;
		font-family: 'Cinzel', serif;
	}

	.channel-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.status-dot {
		width: 10px;
		height: 10px;
		background: #44ff44;
		border-radius: 50%;
		box-shadow: 0 0 8px #44ff44;
	}

	h2 {
		margin: 0;
		font-size: 1.2rem;
		text-transform: uppercase;
		letter-spacing: 1px;
	}

	.close-btn {
		background: none;
		border: none;
		color: #f7e7c4;
		font-size: 1.5rem;
		cursor: pointer;
		transition: color 0.2s;
	}

	.close-btn:hover {
		color: #8c2a3e;
	}

	.message-container {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		background-image: radial-gradient(#5b4a3c 0.5px, transparent 0.5px);
		background-size: 20px 20px;
		background-color: #fffdf5;
	}

	.message-wrapper {
		display: flex;
		width: 100%;
	}

	.message-wrapper.me { justify-content: flex-end; }
	.message-wrapper.them { justify-content: flex-start; }

	.bubble {
		max-width: 80%;
		padding: 4px 12px;
		border-radius: 12px;
		position: relative;
		font-family: 'IM Fell English', serif;
		font-size: 1.1rem;
		border: 2px solid #5b4a3c;
		box-shadow: 2px 2px 0px rgba(0,0,0,0.1);
	}

	.me .bubble {
		background: #f7e7c4;
		color: #3e322b;
		border-bottom-right-radius: 2px;
	}

	.them .bubble {
		background: #ffffff;
		color: #3e322b;
		border-bottom-left-radius: 2px;
	}

	.sender {
		display: block;
		font-size: 0.8rem;
		font-weight: bold;
		color: #8c2a3e;
		margin-bottom: 4px;
		text-transform: uppercase;
	}

	.timestamp {
		display: block;
		font-size: 0.7rem;
		text-align: right;
		margin-top: 5px;
		opacity: 0.7;
	}

	footer {
		padding: 15px;
		background: #3e322b;
		border-top: 3px solid #5b4a3c;
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.input-wrapper { flex: 1; }
	
	:global(.input-wrapper input) {
		margin: 0 !important;
	}

	.send-wrapper :global(button) {
		margin: 0 !important;
		padding: 8px 20px !important;
	}
</style>