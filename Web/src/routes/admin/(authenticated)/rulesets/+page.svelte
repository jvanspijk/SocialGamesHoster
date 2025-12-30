<script lang="ts">
	import BackLink from '$lib/components/BackLink.svelte';
    import type { PageProps } from './$types';

    let { data }: PageProps = $props();
</script>

    <div class="page-header">
        <BackLink href="/admin" pageName="Admin Overview"></BackLink>
        <h1>Manage & create rulesets</h1>
    </div>
