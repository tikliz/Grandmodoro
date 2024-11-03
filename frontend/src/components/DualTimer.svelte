<script lang="ts">
    import { onMount } from "svelte";
    import { onDestroy } from "svelte";
    import { EventsOn, EventsOff } from "../../wailsjs/runtime";
    import {
        StartDualTimer,
        StopDualTimer,
        GetRunningPrograms,
        CheckFocused,
        PauseDualTimer,
        ResumeDualTimer,
    } from "../../wailsjs/go/main/App";
    import { slide } from "svelte/transition";
    import { main } from "../../wailsjs/go/models";

    const FRACTION = 16.67;

    export let workDuration = 60000;
    export let restDuration = 60000;
    export let topBarDuration = ((FRACTION * 3) / 100) * workDuration;
    export let bottomBarDuration = workDuration - topBarDuration;
    export let topBarTick = 500;
    export let bottomBarTick = 1000;
    let topBarValue = topBarDuration;
    let bottomBarValue = bottomBarDuration;
    let workInputMinutes = Math.floor(workDuration / 60000);
    let restInputMinutes = Math.floor(workDuration / 60000);
    let sliderValue = FRACTION * 3;
    let working = true;
    let isPaused = false;

    let topBarWidth = 100;
    let bottomBarWidth = 100;
    let refillInterval;
    let topBarInterval;

    let isDropdownOpen = false;
    let showProgramDropdown = false;

    let programs: main.ProgramInfo[] = [];
    let filteredPrograms: main.ProgramInfo[] = [];

    let programFilterText = "";
    let selectedPrograms: string[] = [];

    onMount(async () => {
        await StartDualTimer(topBarTick, bottomBarTick);
        updateDurations(sliderValue);
        togglePauseResume();

        EventsOn("topBarTick", async () => {
            if (working) {
                if (
                    selectedPrograms.length > 0 &&
                    (await CheckFocused(selectedPrograms)) == false
                ) {
                    return;
                }
                topBarValue -= topBarTick;
                if (topBarValue <= 0) {
                    topBarValue = 0;
                    if (bottomBarValue <= 0) {
                        working = false;
                    }
                }
            } else {
                if (topBarValue >= topBarDuration) {
                    topBarValue = topBarDuration;
                    if (bottomBarValue >= bottomBarDuration) {
                        bottomBarValue = bottomBarDuration;
                        working = true;
                    }
                    return;
                }
                let temp = parseInt(
                    (bottomBarValue % topBarDuration).toString()[0],
                );
                if (temp >= 9) {
                    const refillAmount = bottomBarValue;
                    const maxAddableToTop = topBarDuration - topBarValue;

                    if (refillAmount <= maxAddableToTop) {
                        topBarValue += refillAmount;
                        bottomBarValue = 0;
                    } else {
                        topBarValue = topBarDuration; // Fill the top bar completely
                        const excess = refillAmount - maxAddableToTop;
                        bottomBarValue = Math.max(0, bottomBarValue - excess);
                    }
                }
            }

            topBarWidth = (topBarValue / topBarDuration) * 100;
        });

        EventsOn("bottomBarTick", () => {
            if (working) {
                if (topBarValue < topBarDuration && bottomBarValue > 0) {
                    const refillAmount = bottomBarTick;
                    const maxAddableToTop = topBarDuration - topBarValue;

                    if (refillAmount <= maxAddableToTop) {
                        topBarValue += refillAmount;
                        bottomBarValue -= refillAmount;
                    } else {
                        topBarValue = topBarDuration;
                        const excess = refillAmount - maxAddableToTop;
                        bottomBarValue = Math.max(0, bottomBarValue - excess);
                    }

                    topBarWidth = (topBarValue / topBarDuration) * 100;
                    bottomBarWidth = (bottomBarValue / bottomBarDuration) * 100;

                    if (bottomBarWidth < 0) bottomBarWidth = 0;
                }
            } else {
                if (bottomBarValue < bottomBarDuration) {
                    const refillAmount = bottomBarTick;
                    const maxAddableToBottom =
                        bottomBarDuration - bottomBarValue;

                    if (refillAmount <= maxAddableToBottom) {
                        bottomBarValue += refillAmount;
                    } else {
                        bottomBarValue = bottomBarDuration;
                    }

                    bottomBarWidth = (bottomBarValue / bottomBarDuration) * 100;

                    if (bottomBarWidth > 100) bottomBarWidth = 100;
                }
            }
        });
    });

    onDestroy(() => {
        StopDualTimer();
        EventsOff("topBarTick");
        EventsOff("bottomBarTick");
    });

    const toggleProgramSelection = (program: main.ProgramInfo) => {
        const index = selectedPrograms.findIndex((p) => p === program.Name);
        if (index === -1) {
            selectedPrograms.push(program.Name);
        } else {
            selectedPrograms.splice(index, 1);
        }
        selectedPrograms = selectedPrograms;
    };

    const updateDurations = (value) => {
        if (working) {
            topBarDuration = (value / 100) * workDuration;
            bottomBarDuration = workDuration - topBarDuration;
        } else {
            topBarDuration = (value / 100) * restDuration;
            bottomBarDuration = restDuration - topBarDuration;
        }
    };

    const handleSliderChange = (event) => {
        const value = event.target.value;
        sliderValue = value;
        updateDurations(value);
        resetAll(working);
    };

    const toggleDropdown = () => {
        isDropdownOpen = !isDropdownOpen;
    };

    function togglePauseResume() {
        if (isPaused) {
            ResumeDualTimer();
        } else {
            PauseDualTimer();
        }
        isPaused = !isPaused;
    }

    const updateTotalDurations = () => {
        if (!workInputMinutes || !restInputMinutes) {
            return;
        }
        workDuration = workInputMinutes * 60000;
        restDuration = restInputMinutes * 60000;
        updateDurations(sliderValue);
        // resetAll();
    };

    function resetAll(work = true) {
        topBarValue = topBarDuration;
        bottomBarValue = bottomBarDuration;

        topBarWidth = 100;
        bottomBarWidth = 100;
        working = work;
        updateDurations(sliderValue);
        clearInterval(topBarInterval);
        clearInterval(refillInterval);
        StartDualTimer(topBarTick, bottomBarTick);
    }

    const getFraction = (value) => {
        const topFraction = value / 100;
        const bottomFraction = 1 - topFraction;

        return `${(topFraction * 6).toFixed(0)}/6`;
    };

    function depleteBars() {
        if (!working) {
            topBarValue = topBarDuration;
            bottomBarValue = bottomBarDuration;
        } else {
            topBarValue = 0;
            bottomBarValue = 0;
        }
    }

    $: if (working) {
        updateDurations(sliderValue);
    }

    $: if (selectedPrograms) {
        filteredPrograms = filteredPrograms;
    }

    $: if (programFilterText) {
        if (filteredPrograms.length > 0) {
            filteredPrograms = programs.filter((program) =>
                program.Name.toLowerCase().includes(
                    programFilterText.toLowerCase(),
                ),
            );
        } else {
            filteredPrograms = programs;
        }
    }

    $: sortedPrograms = [...filteredPrograms].sort((a, b) => {
        const aSelected = selectedPrograms.some((p) => p === a.Name);
        const bSelected = selectedPrograms.some((p) => p === b.Name);

        // If a is selected and b is not, put a first
        if (aSelected && !bSelected) return -1;
        // If b is selected and a is not, put b first
        if (!aSelected && bSelected) return 1;
        // Otherwise, keep original order
        return 0;
    });
</script>

<div class="relative items-center p-4 rounded-lg shadow-md bg-slate-800" style="--wails-draggable:drag">
    <div on:click={toggleDropdown} on:keydown class="relative space-y-1">
        <!-- Top fill bar -->
        <div class="w-full h-5 bg-gray-200 rounded-full overflow-hidden">
            <div
                class="h-full bg-green-500 transition-all duration-100 rounded-full"
                style="width: {topBarWidth}%;"
            >
                {#each Array(4) as _, index}
                    <div
                        class="absolute h-5 border-l border-gray-400"
                        style="left: {(index + 1) * 20}%; pointer-events: none;"
                    ></div>
                {/each}
            </div>
        </div>

        <!-- Bottom deplete bar -->
        <div class="w-full h-5 bg-gray-200 rounded-full overflow-hidden">
            <div
                class="h-full bg-red-500 transition-all duration-100 rounded-full"
                style="width: {bottomBarWidth}%;"
            >
                {#each Array(4) as _, index}
                    <div
                        class="absolute h-5 border-l border-gray-400"
                        style="left: {(index + 1) * 20}%; pointer-events: none;"
                    ></div>
                {/each}
            </div>
        </div>
        <div class="flex flex-row">
            <div class="w-64"></div>
            <button
                on:click|stopPropagation={togglePauseResume}
                class="flex items-center justify-center bg-blue-600 text-white px-4 rounded-full w-full"
            >
                {#if isPaused}
                    <!-- Play Icon -->
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                        class="size-4"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z"
                        />
                    </svg>
                {:else}
                    <!-- Pause Icon -->
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                        class="size-4"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M15.75 5.25v13.5m-7.5-13.5v13.5"
                        />
                    </svg>
                {/if}
            </button>
            <div class="w-64"></div>
        </div>
    </div>

    <!-- Timer Options -->
    {#if isDropdownOpen}
        <div
            class="absolute left-0 right-0 mx-auto mt-2 p-4 bg-white rounded shadow-lg w-full max-w-xs max-h-[315px] overflow-y-auto"
        >
            <div class="flex flex-row content-between mt-4">
                <div>
                    <label
                        for="work-duration-input"
                        class="block text-sm font-medium text-gray-700"
                        >Work Duration (minutes)</label
                    >
                    <input
                        id="work-duration-input"
                        type="number"
                        min="1"
                        step="1"
                        bind:value={workInputMinutes}
                        class="text-center mt-1 block w-full py-1 border text-gray-700 rounded"
                        on:input={updateTotalDurations}
                        draggable="false"
                    />
                </div>
                <div>
                    <label
                        for="rest-duration-input"
                        class="block text-sm font-medium text-gray-700"
                        >Rest Duration (minutes)</label
                    >
                    <input
                        id="rest-duration-input"
                        type="number"
                        min="1"
                        step="1"
                        bind:value={restInputMinutes}
                        class="text-center mt-1 block w-full py-1 border text-gray-700 rounded"
                        on:input={updateTotalDurations}
                    />
                </div>
            </div>
            <p class="mt-2 text-sm text-gray-600">
                Current Total Duration: {restInputMinutes + workInputMinutes} minutes
            </p>
            <label
                for="duration-slider"
                class="block text-sm font-medium text-gray-700"
                >Adjust Visual Speed</label
            >
            <input
                id="slider"
                type="range"
                min={FRACTION}
                max="100.02"
                step={FRACTION}
                bind:value={sliderValue}
                class="w-full"
                on:input={handleSliderChange}
            />
            <div class="flex justify-between text-sm text-gray-600">
                <span>1 min</span>
                <span>{getFraction(sliderValue)}</span>
                <span>{workDuration / 60000} min</span>
            </div>
            <div class="flex flex-row">
                <button
                    on:click={depleteBars}
                    class="mt-2 w-full bg-gray-300 rounded-lg appearance-none cursor-pointer"
                    >DEPLETE</button
                >
                <button
                    on:click={() => resetAll()}
                    class="mt-2 w-full bg-gray-300 rounded-lg appearance-none cursor-pointer"
                    >RESET TIMER</button
                >
            </div>
            <!-- Nested dropdown for running programs -->
            <button
                on:click={async () => {
                    showProgramDropdown = !showProgramDropdown;
                    if (showProgramDropdown) {
                        programs = await GetRunningPrograms();
                        filteredPrograms = programs;
                    }
                }}
                class="mt-2 p-2 bg-gray-100 rounded w-full text-gray-700"
            >
                Select Running Program
            </button>

            {#if showProgramDropdown}
                <div
                    class="mt-2 w-full max-h-24 bg-white border border-gray-300 rounded shadow overflow-y-auto"
                >
                    <input
                        type="text"
                        placeholder="Filter programs..."
                        bind:value={programFilterText}
                        class="block w-full p-2 border rounded text-gray-700"
                    />

                    <div
                        class="absolute z-10 w-[90%] bg-white border rounded shadow-lg"
                    >
                        {#each sortedPrograms as program}
                            <button
                                class={`block p-2 text-sm w-full text-left 
                ${selectedPrograms.some((p) => p === program.Name) ? "bg-blue-500 text-white" : "hover:bg-gray-200 text-gray-700"}`}
                                on:click={() => toggleProgramSelection(program)}
                            >
                                {program.Name}
                            </button>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>
    {/if}
</div>
