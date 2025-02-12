<script>
  import { onMount } from 'svelte';
  import { getContext } from 'svelte';

  let queryParams = new URLSearchParams(window.location.search);

  let trace_turn_policy = queryParams.get('trace_turn_policy') || 'minority';
  let trace_turd_size = Number(queryParams.get('trace_turd_size')) || 10000;
  let trace_alpha_max = Number(queryParams.get('trace_alpha_max')) || 1;
  let trace_curve_optimization_tolerance = Number(queryParams.get('trace_curve_optimization_tolerance')) || 0;
  let black_level = Number(queryParams.get('black_level')) || 0;
  let invert = queryParams.get('invert') === 'true';
  let ma3_scribble_name = queryParams.get('ma3_scribble_name') || '';
  let ma3_scribble_stroke_thickness = Number(queryParams.get('ma3_scribble_stroke_thickness')) || 0;
  let ma3_scribble_stroke_color = queryParams.get('ma3_scribble_stroke_color') || 'rgba(0,0,0,1)';
</script>

<main>
  <label>
    Trace Turn Policy:
    <select bind:value={trace_turn_policy}>
      <option value="black">black</option>
      <option value="white">white</option>
      <option value="right">right</option>
      <option value="left">left</option>
      <option value="minority">minority</option>
      <option value="majority">majority</option>
      <option value="random">random</option>
    </select>
  </label>

  <label>
    Trace Turd Size: {trace_turd_size}
    <input type="number" min="0" bind:value={trace_turd_size}>
    <input type="range" min="0" max="50000" step="1" bind:value={trace_turd_size}>
  </label>

  <label>
    Trace Alpha Max: {trace_alpha_max}
    <input type="number" min="0" max="1.4" step="0.01" bind:value={trace_alpha_max}>
    <input type="range" min="0" max="1.4" step="0.01" bind:value={trace_alpha_max}>
  </label>

  <label>
    Trace Curve Optimization Tolerance: {trace_curve_optimization_tolerance}
    <input type="number" min="0" bind:value={trace_curve_optimization_tolerance}>
    <input type="range" min="0" max="10" step="0.1" bind:value={trace_curve_optimization_tolerance}>
  </label>

  <label>
    Black Level: {black_level}
    <input type="number" min="0" max="1" step="0.01" bind:value={black_level}>
    <input type="range" min="0" max="1" step="0.01" bind:value={black_level}>
  </label>

  <label>
    Invert:
    <input type="checkbox" bind:checked={invert}>
  </label>

  <label>
    MA3 Scribble Name:
    <input type="text" bind:value={ma3_scribble_name}>
  </label>

  <label>
    MA3 Scribble Stroke Thickness: {ma3_scribble_stroke_thickness}
    <input type="number" min="0" max="10" step="0.1" bind:value={ma3_scribble_stroke_thickness}>
    <input type="range" min="0" max="10" step="0.1" bind:value={ma3_scribble_stroke_thickness}>
  </label>

  <label>
    MA3 Scribble Stroke Color:
    <input type="color" bind:value={ma3_scribble_stroke_color}>
  </label>
</main>
