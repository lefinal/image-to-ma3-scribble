<script lang="ts">
  interface Params {
    preprocess_transparency_replacement_color: string;
    preprocess_blur_radius: number;
    trace_turn_policy: string;
    trace_turd_size: number;
    trace_alpha_max: number;
    trace_curve_optimization_tolerance: number;
    black_level: number;
    invert: boolean;
    ma3_scribble_name: string;
    ma3_scribble_stroke_thickness: number;
    ma3_scribble_stroke_color: string;
  }

  let queryParams = new URLSearchParams(window.location.search);

  const maxCurvesInMA3Scribble = 129;

  let params: Params = {
    preprocess_transparency_replacement_color: queryParams.get('preprocess_transparency_replacement_color') || '#ffffff',
    preprocess_blur_radius: Number(queryParams.get('preprocess_blur_radius')) || 0.0,
    trace_turn_policy: queryParams.get('trace_turn_policy') || 'minority',
    trace_turd_size: Number(queryParams.get('trace_turd_size')) || 10000,
    trace_alpha_max: Number(queryParams.get('trace_alpha_max')) || 1,
    trace_curve_optimization_tolerance: Number(queryParams.get('trace_curve_optimization_tolerance')) || 0.5,
    black_level: Number(queryParams.get('black_level')) || 0.15,
    invert: queryParams.get('invert') === 'true',
    ma3_scribble_name: queryParams.get('ma3_scribble_name') || '',
    ma3_scribble_stroke_thickness: Number(queryParams.get('ma3_scribble_stroke_thickness')) || 2,
    ma3_scribble_stroke_color: queryParams.get('ma3_scribble_stroke_color') || '#ffffff',
  };

  const serviceBaseUrl = "https://la-solutions.one/apps/image-to-ma3-scribble"
  // const serviceBaseUrl = 'http://localhost:8001';

  function updateQueryParamsInUrl(params: Params) {
    const url = new URL(window.location.href);

    // Set new query params
    Object.keys(params).forEach(key => {
      url.searchParams.set(key, params[key]);
    });

    // Update the URL without reloading
    window.history.replaceState({}, '', url);
  }

  let file;
  let previewImage = '';
  let timeout;

  // Trigger preview after a delay when any input is changed
  function updateParams() {
    if (timeout) {
      clearTimeout(timeout); // Clear any previous timeout
    }

    timeout = setTimeout(() => {
      track('update_params', { params: params });
      preview();
    }, 200); // Delay for 200 milliseconds after the last change
  }

  // Manually bind the inputs to this function to trigger preview update
  function handleFileUpload(event) {
    file = event.target.files[0];
    if (file && file.type !== 'image/png') {
      track('invalid_file_type', { file_type: file.type });
      alert('Only PNG files are allowed.');
      file = null;
    }
    track('file_upload', { file_type: file ? file.type : null });
    updateParams();  // Trigger preview on file upload
  }

  function prepareParamsForQuery(): URLSearchParams {
    const clonedParams = { ...params };
    // Convert the transparency replacement color to an 8-character hex value
    if (clonedParams.preprocess_transparency_replacement_color.length < 8) {
      clonedParams.preprocess_transparency_replacement_color += 'ff';
    }
    // Convert the stroke color to an 8-character hex value
    if (clonedParams.ma3_scribble_stroke_color.length < 8) {
      clonedParams.ma3_scribble_stroke_color += 'ff';
    }
    let urlQueryParams = new URLSearchParams(clonedParams);
    return urlQueryParams;
  }

  async function preview() {
    if (!file) {
      return;
    }

    track('preview', { params: params });

    updateQueryParamsInUrl(params);

    const arrayBuffer = await file.arrayBuffer();

    const urlQueryParams = prepareParamsForQuery();

    try {
      const response = await fetch(`${ serviceBaseUrl }/api/v1/png-to-ma3-scribble/preview?${ urlQueryParams.toString() }`, {
        method: 'POST',
        headers: {
          'Content-Type': 'image/png',  // Indicate that you're sending a PNG file
        },
        body: arrayBuffer,  // Send the file as raw binary data
      });

      if (response.ok) {
        let svgText = await response.text();

        // Use DOMParser to parse the SVG string into an XML document
        const parser = new DOMParser();
        const svgDoc = parser.parseFromString(svgText, 'image/svg+xml');

        // Remove width and height attributes from the <svg> element
        const svgElement = svgDoc.querySelector('svg');
        if (svgElement) {
          svgElement.removeAttribute('width');
          svgElement.removeAttribute('height');
        }

        // Serialize the modified SVG back to a string
        const serializer = new XMLSerializer();
        const modifiedSvgText = serializer.serializeToString(svgDoc);

        // Insert the modified SVG into the container
        previewImage = modifiedSvgText;
      } else {
        const errorContent = await response.text();  // Retrieve the response content
        console.error('Error Response:', errorContent);
        alert(`Error: ${ errorContent }`);
      }
    } catch (error) {
      console.error('Error:', error);
      alert('Request failed. Check console for details.');
    }
  }

  async function download() {
    if (!file) {
      return;
    }

    track('download', { params: params });

    updateQueryParamsInUrl(params);

    const arrayBuffer = await file.arrayBuffer();

    const urlQueryParams = prepareParamsForQuery();

    try {
      const response = await fetch(`${ serviceBaseUrl }/api/v1/png-to-ma3-scribble?${ urlQueryParams.toString() }`, {
        method: 'POST',
        headers: {
          'Content-Type': 'image/png',  // Indicate that you're sending a PNG file
        },
        body: arrayBuffer,  // Send the file as raw binary data
      });

      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');

        a.href = url;
        a.download = `${ params.ma3_scribble_name || 'scribble' }.xml`; // Set filename, fallback to 'scribble.xml'
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      } else {
        const errorContent = await response.text();  // Retrieve the response content
        console.error('Error Response:', errorContent);
        alert(`Error: ${ errorContent }`);
      }
    } catch (error) {
      console.error('Error:', error);
      alert('Request failed. Check console for details.');
    }
  }

  /**
   * Track an event.
   * @param eventName
   * @param eventDetails
   */
  export function track(eventName: string, eventDetails?: object): void {
    const umami = window.umami;
    if (umami) {
      // @ts-expect-error Global variable that is created via root +layout.svelte.
      umami.track(eventName, eventDetails);
    }
  }

  function countPathSegments(svgString: string) {
    // Create a temporary DOM parser
    const parser = new DOMParser();
    const doc = parser.parseFromString(svgString, 'image/svg+xml');

    // Get all <path> elements
    const paths = doc.querySelectorAll('path');
    let segmentCount = 0;

    // Define regex to match path commands
    const pathRegex = /([CSQTLHVcsqtlhv])([^CSQTLHVcsqtlhv]*)/g;

    paths.forEach(path => {
      const d = path.getAttribute('d');
      if (d) {
        let match;
        while ((match = pathRegex.exec(d)) !== null) {
          const command = match[1];  // The command letter
          const values = match[2].trim().split(/[\s,]+/).map(Number).filter(n => !isNaN(n));

          switch (command) {
            // Bézier Curves
            case 'C':
            case 'c': // Cubic Bézier (6 numbers per segment)
              segmentCount += values.length / 6;
              break;
            case 'S':
            case 's': // Smooth cubic Bézier (4 numbers per segment)
              segmentCount += values.length / 4;
              break;
            case 'Q':
            case 'q': // Quadratic Bézier (4 numbers per segment)
              segmentCount += values.length / 4;
              break;
            case 'T':
            case 't': // Smooth quadratic Bézier (2 numbers per segment)
              segmentCount += values.length / 2;
              break;

            // Lines
            case 'L':
            case 'l': // Line (2 numbers per segment)
              segmentCount += values.length / 2;
              break;
            case 'H':
            case 'h': // Horizontal Line (1 number per segment)
              segmentCount += values.length;
              break;
            case 'V':
            case 'v': // Vertical Line (1 number per segment)
              segmentCount += values.length;
              break;
          }
        }
      }
    });

    return Math.floor(segmentCount);
  }

  function trackFocusVersion() {
    track('focus_version');
  }

  function trackFocusAuthor() {
    track('focus_author');
  }

  function trackFocusLicense() {
    track('focus_license');
  }
</script>

<svelte:head>
  <script async src="https://analytics.la-solutions.one/umami" data-website-id="c4ba42e4-9a92-4498-b846-a689ad63fb17"
          data-domains="localho.st"></script>
  <script async data-website-id="167fcd8d-55b2-4c7d-ab56-0fe66d1b4036"
          data-domains="la-solutions.one,www.la-solutions.one"
          src="https://analytics.la-solutions.one/umami"></script>
</svelte:head>

<style>

    main {
        max-width: 100vw;
        /*height: 100vh;*/
        background: #222;
        padding: 0;
        margin: 0;
        box-sizing: content-box;
        font-family: sans-serif;

        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        flex-grow: 1;
        align-content: center;
        justify-content: center;
        gap: 1em;
        padding-bottom: 40px;
        min-height: 100vh;
    }

    .column {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        flex-shrink: 1;
        align-content: start;
        justify-content: center;
        gap: 1em;
        padding-bottom: 40px;
    }

    .widget {
        width: 500px;
        max-width: 600px;
        height: fit-content;
        background: #333;
        padding: 20px;
        border-radius: 10px;
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
        color: #eee;
    }

    .params {
        display: flex;
        flex-direction: column;
        gap: .2em;
    }

    .preview {
        display: flex;
        flex-direction: column;
        gap: 1em;
    }

    .hint {
        color: #666;
    }

    label {
        display: flex;
        flex-direction: column;
        margin: 10px 0;
        font-size: 14px;
    }

    input, select, button {
        padding: 8px;
        margin-top: .5em;
        border: 1px solid #aaa;
        border-radius: 5px;
        color: #eee;
        background: #4a4a4a;
    }

    input[type="range"] {
        width: 80%;
        margin-top: 5px;
    }

    input[type="checkbox"] {
        width: fit-content;
        vertical-align: bottom;
    }

    label.checkbox {
        display: block;
    }

    .help {
        display: inline-block;
        width: 1.5em;
        height: 1.5em;
        margin-left: .5em;
        border-radius: 50%;
        background-color: #16589b;
        color: white;
        text-align: center;
        font-size: 1em;
        font-weight: bold;
        cursor: pointer;
        position: relative;
        vertical-align: middle;
    }

    .help-tooltip {
        visibility: hidden;
        background-color: black;
        color: white;
        text-align: center;
        padding: 5px;
        border-radius: 5px;
        position: absolute;
        left: 50%;
        bottom: 120%;
        width: 20em;
        max-width: 30vw;
        word-wrap: normal;
        transform: translateX(-50%);
        white-space: wrap;
        font-size: 12px;
        opacity: 0;
        transition: opacity 0.2s;
    }

    .help:hover .help-tooltip {
        visibility: visible;
        opacity: 1;
    }

    .actions {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        gap: 1em;
    }

    button {
        flex-grow: 1;
        background-color: #007bff;
        color: white;
        border: none;
        cursor: pointer;
    }

    button:hover {
        background-color: #0056b3;
    }

    button:disabled {
        background-color: #444444;
    }

    button:disabled:hover {
        background-color: #444444;
    }

    .preview-container {
        border: 1px solid #777;
        margin-top: 20px;
        text-align: center;
        width: 100%;
        max-width: 600px;
        height: auto;
        object-fit: contain;
    }

    .preview-container svg {
        width: 100%;
        height: 100%;
        object-fit: contain;
    }

    h1 {
        margin-bottom: .3em;
    }

    .version {
        font-size: .5em;
        margin-left: .2em;
    }

    .version, .author {
        color: #aaa;
    }

    a, a:focus, a:active {
        color: inherit;
    }

    a:hover {
        color: #4a94dd;
    }
</style>

<main>
  <div class="column">
    <div class="widget">
      <h1>PNG to MA3 Scribble <span class="version">v{PKG.version}</span></h1>
      <div class="author">by <a href="https://la-solutions.one" on:focus={trackFocusAuthor}>Lennart Altenhof</a></div>

      <div class="params">
        <h3>Image upload</h3>
        <label>
          Upload Image or drop below (.png)
          <input type="file" accept="image/png" on:change={handleFileUpload}>
        </label>
        <div class="hint">
          Images will get converted to a black-white bitmap before tracing.
          Use an already black-white image to get the best results.
        </div>
      </div>

      <div class="params">
        <h3>Preprocessing</h3>
        <label>
          <span>
            Transparency Replacement Color
            <span class="help">?
              <span class="help-tooltip">Transparency in the PNG will be replaced with this color as tracing only works without transparency.</span>
            </span>
          </span>
          <input type="color" bind:value={params.preprocess_transparency_replacement_color} on:input={updateParams}>
        </label>

        <label>
      <span>
        Blur Radius
        <span class="help">?
          <span class="help-tooltip">Radius for blurring the image. Useful for reducing the number of curves.</span>
        </span>
      </span>
          <input type="number" min="0" max="4000" bind:value={params.preprocess_blur_radius}
                 on:input={updateParams}>
          <input type="range" min="0" max="20" step=".01" bind:value={params.preprocess_blur_radius}
                 on:input={updateParams}>
        </label>
      </div>
    </div>

    <div class="widget">
      <div class="params">
        <h3>Tracing</h3>
        <label>
      <span>
        Turn Policy
        <span class="help">?
          <span class="help-tooltip">Defines how corners are handled in tracing.</span>
        </span>
      </span>
          <select bind:value={params.trace_turn_policy} on:change={updateParams}>
            <option value="black">Black</option>
            <option value="white">White</option>
            <option value="right">Right</option>
            <option value="left">Left</option>
            <option value="minority">Minority</option>
            <option value="majority">Majority</option>
            <option value="random">Random</option>
          </select>
        </label>

        <label>
      <span>
        Speckle Suppression
        <span class="help">?
          <span class="help-tooltip">Suppress small speckles of up to this size.</span>
        </span>
      </span>
          <input type="number" max="100000000" bind:value={params.trace_turd_size} on:input={updateParams}>
          <input type="range" min="1000" max="20000" step="100" bind:value={params.trace_turd_size}
                 on:input={updateParams}>
        </label>

        <label>
      <span>
        Corner To Curve Optimization
        <span class="help">?
          <span class="help-tooltip">Threshold for detecting corners. If smaller, more sharp corners will be produced instead of curves.</span>
        </span>
      </span>
          <input type="number" step="0.1" min="0" max="1.4" bind:value={params.trace_alpha_max} on:input={updateParams}>
          <input type="range" min="0" max="1.4" step="0.01" bind:value={params.trace_alpha_max} on:input={updateParams}>
        </label>

        <label>
      <span>
        Curve Optimization Tolerance
        <span class="help">?
          <span class="help-tooltip">Optimization tolerance for curve optimization. The larger the value the more curves will be joined together.</span>
        </span>
      </span>
          <input type="number" min="0" max="100000000" bind:value={params.trace_curve_optimization_tolerance}
                 on:input={updateParams}>
          <input type="range" min="0" max="20" step=".01" bind:value={params.trace_curve_optimization_tolerance}
                 on:input={updateParams}>
        </label>

        <label>
      <span>
        Black Level Threshold
        <span class="help">?
          <span class="help-tooltip">Threshold for converting the input image to a black-white bitmap.</span>
        </span>
      </span>
          <input type="number" min="0" max="1" step="0.01" bind:value={params.black_level} on:input={updateParams}>
          <input type="range" min="0" max="1" step="0.01" bind:value={params.black_level} on:input={updateParams}>
        </label>

        <label class="checkbox">
          Invert Image
          <input type="checkbox" bind:checked={params.invert} on:change={updateParams}>
        </label>
      </div>
    </div>
  </div>

  <div class="column">
    <div class="widget">
      <div class="params">
        <h3>Scribble</h3>

        <label>
          Scribble Name
          <input type="text" bind:value={params.ma3_scribble_name} on:input={updateParams}>
        </label>

        <label>
          Stroke Thickness
          <input type="number" min="0" max="10" bind:value={params.ma3_scribble_stroke_thickness}
                 on:input={updateParams}>
          <input type="range" min="0" max="10" step="0.1" bind:value={params.ma3_scribble_stroke_thickness}
                 on:input={updateParams}>
        </label>

        <label>
          Stroke Color
          <input type="color" bind:value={params.ma3_scribble_stroke_color} on:input={updateParams}>
        </label>
      </div>

      <div class="actions">
        <button on:click={preview} disabled={!file}>Preview Path</button>
        <button on:click={download} disabled={!file}>Build Scribble</button>
      </div>
    </div>

    <div class="widget">
      <div class="preview">
        <h3>Path Preview</h3>
        <div class="preview-container" contenteditable="true" bind:innerHTML={previewImage}></div>
        {#if previewImage}
          <div>
            Curves: {countPathSegments(previewImage)}/{maxCurvesInMA3Scribble}
            <span class="help">?
              <span class="help-tooltip">
                The maximum number of curves in MA3 scribbles is {maxCurvesInMA3Scribble}.
                If you have too many curves, try increasing curve optimization tolerance, blur or just use a simpler image.
              </span>
            </span>
          </div>
          <div class="hint">If you do not see any output try to adjust the black level or upload a black and white
            image.
          </div>
        {/if}
        <div>
          Click <i>Build Scribble</i> to download.
          Place the resulting <code>.xml</code> file in your scribbles folder, located in <i>C:\Program Data\&lt;MA
          directory&gt;\gma3_library\scribbles</i>.
          Then, in MA3, click <code>Menu</code>→<code>Show Creator</code>→<code>Import</code>→<code>Scribbles</code> and
          import your scribble.
        </div>
        <div></div>
        <div class="hint">
          This app was primarily stitched together within a day for my personal use to get some annoying tasks done.
          Its code quality is light years behind what I would normally build.
          However, I thought it might be useful for others as well.
        </div>
        <div class="hint">
          If stuff doesn't work as expected or you have any feedback, please let me know.
          Keep in mind that my goal was not to create a 1:1 representation, but to convert simple images to a format
          that
          can be used in MA3 without automatized cursor movement.
        </div>
        <div class="hint">
          Because the program is licensed free of charge, there is no warranty for the program, to the extent permitted
          by
          applicable law.
          See license details <a href="https://github.com/lefinal/image-to-ma3-scribble"
                                 on:focus={trackFocusLicense}>here</a>.
        </div>
      </div>
    </div>
  </div>

</main>
