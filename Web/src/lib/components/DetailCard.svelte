<script lang="ts">
    let { title, content } = $props();
</script>

<div class="detail-card">
    <h2>{ title }</h2>
    <p>{ content }</p>                   
</div>

<style>
    .detail-card {
        background-color: #f7f3e8;
        border: 1px solid #d4c5a1;
        padding: 20px;
        box-shadow: 4px 4px 10px rgba(0, 0, 0, 0.1);
        border-radius: 4px;
    }

    .detail-card h2 {
        font-family: 'Cinzel', serif;
        color: #a62a2a;
        margin-top: 0;
        margin-bottom: 10px;
        border-bottom: 2px solid #a62a2a;
        padding-bottom: 5px;
        font-size: 1.25em;
    }
</style>