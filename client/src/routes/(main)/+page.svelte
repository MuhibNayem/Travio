<script lang="ts">
	import HeroLanding from "$lib/components/landing/HeroLanding.svelte";
	import StatsStrip from "$lib/components/landing/StatsStrip.svelte";
	import FeatureGrid from "$lib/components/blocks/FeatureGrid.svelte";
	import TrustedPartners from "$lib/components/landing/TrustedPartners.svelte";
	import TestimonialCarousel from "$lib/components/landing/TestimonialCarousel.svelte";
	import AppDownloadSection from "$lib/components/landing/AppDownloadSection.svelte";
</script>

<svelte:head>
	<title>TicketNation - Your Nationwide Travel Companion</title>
	<meta
		name="description"
		content="Book train, bus, launch, and air tickets seamlessly across the nation. Trusted by millions."
	/>
</svelte:head>

<div id="top">
	<HeroLanding />
	<StatsStrip />

	<div id="features" class="py-20">
		<FeatureGrid />
	</div>

	<TrustedPartners />
	<TestimonialCarousel />
	<AppDownloadSection />
</div>
