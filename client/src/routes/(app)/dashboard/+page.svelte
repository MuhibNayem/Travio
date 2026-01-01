<script lang="ts">
    import SearchHero from "$lib/components/blocks/SearchHero.svelte";
    import CategoryGrid from "$lib/components/blocks/CategoryGrid.svelte";
    import TrendingSection from "$lib/components/blocks/TrendingSection.svelte";
    import PromoBanner from "$lib/components/blocks/PromoBanner.svelte";
    import FeatureGrid from "$lib/components/blocks/FeatureGrid.svelte";
</script>

<svelte:head>
    <title>TicketNation - Nationwide Ticketing Platform</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />

    <link
        href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;700;800&display=swap"
        rel="stylesheet"
    />
    <link
        href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@100..700&display=swap"
        rel="stylesheet"
    />
</svelte:head>

<SearchHero />
<CategoryGrid />
<TrendingSection />
<PromoBanner />
<FeatureGrid />
