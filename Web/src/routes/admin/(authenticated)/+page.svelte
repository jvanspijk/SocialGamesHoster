<script lang="ts">
    import NavBlock from "$lib/components/NavBlock.svelte";
</script>

<div class="container">
    <h1>Admin Console</h1>
    <h3>With great power comes great responsibility.</h3>

    <section class="nav-section">
        <h2>Game sessions</h2>
        <NavBlock 
            href="/admin/games" 
            title="Active Games" 
            description="Manage active games."
            icon="⚔️"
        />
        <NavBlock 
            href="/admin/games/new" 
            title="Start New Game" 
            description="Start a new game."
            icon="🔔"
        />
    </section>

    <section class="nav-section">
        <h2>Rulesets</h2>
        <NavBlock 
            href="/admin/rulesets" 
            title="Manage Rulesets" 
            description="Define the roles, abilities, and laws that govern every game."
            icon="📜"
        />
    </section>
</div>

<style>
    .nav-section {
        width: 100%;
        margin-bottom: 2rem;
    }

    h2 {
        text-align: left;
        font-size: 1.2rem;
        border-bottom: 2px solid #5b4a3c;
        margin-top: 1.5rem;
    }
</style>