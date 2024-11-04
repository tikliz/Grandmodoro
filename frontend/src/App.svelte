<script lang="ts">
    import DualTimer from "./components/DualTimer.svelte";


</script>

<main>
  <DualTimer workValue={.2} workDuration={25 * 60000} restDuration={5 * 60000} />
</main>
