<script lang="ts">
    let { href, title, description, icon = "📜" } = $props();
</script>

<a {href} class="nav-block">
    <div class="icon">{icon}</div>
    <div class="text-content">
        <h3>{title}</h3>
        <p>{description}</p>
    </div>
</a>

<style>
    .nav-block {
        display: flex;
        align-items: center;
        text-decoration: none;
        background-color: #fcf5e5;
        border: 2px solid #5b4a3c;
        padding: 1.5rem;
        margin: 1rem 0;
        transition: all 0.2s ease-in-out;
        box-shadow: 3px 3px 0px #5b4a3c;
        cursor: pointer;
        text-align: left;
    }

    .nav-block:hover {
        transform: translate(-2px, -2px);
        box-shadow: 6px 6px 0px #a62a2a;
        background-color: #fffcf5;
        border-color: #a62a2a;
    }

    .icon {
        font-size: 2.5rem;
        margin-right: 1.5rem;
        filter: sepia(1);
    }

    h3 {
        margin: 0;
        color: #a62a2a !important;
        font-family: 'Cinzel', serif;
        font-style: normal !important;
        text-transform: uppercase;
        font-size: 1.2rem;
    }

    p {
        margin: 4px 0 0 0;
        color: #5b4a3c;
        font-size: 0.95rem;
        line-height: 1.2;
    }
</style>