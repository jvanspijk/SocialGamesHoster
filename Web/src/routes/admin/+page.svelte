<script lang="ts">
    import type { PageProps } from './$types';

    let { data }: PageProps = $props();
</script>

<header>
    <nav>
        <a href="admin/games/start">Start Game</a>
        |
        <a href="admin/rulesets/create">Create Ruleset</a>
    </nav>
</header>