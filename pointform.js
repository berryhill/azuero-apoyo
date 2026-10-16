// Formulario compartido de puntos (público y administración): «¿Dónde estás?», «¿Con quién estás?»,
// «¿Qué necesitas?» y, solo para administración, los campos internos.
(function () {
  const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
  const svg = path => `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${path}</svg>`;
  const icons = {
    pin: svg('<path d="M12 21s-7-6.2-7-11a7 7 0 0 1 14 0c0 4.8-7 11-7 11z"/><circle cx="12" cy="10" r="2.5"/>'),
    toilet: svg('<path d="M7 3h4v6H7z"/><path d="M5 9h14a6 6 0 0 1-6 6h-2a6 6 0 0 1-6-6Z"/><path d="M9 15l-1 6h8l-1-6"/>'),
    trash: svg('<path d="M4 7h16"/><path d="M10 3h4l1 4H9z"/><path d="M6 7l1 13h10l1-13"/><path d="M10 11v6M14 11v6"/>'),
    spray: svg('<path d="M9 10h6v11H9z"/><path d="M10 10V6h4v4"/><path d="M14 6h3"/><path d="M19 3v.01M21 5v.01M19 7v.01"/>'),
    people: svg('<circle cx="9" cy="8" r="3"/><path d="M3 20a6 6 0 0 1 12 0"/><circle cx="17" cy="9" r="2.5"/><path d="M15.5 14.2A5 5 0 0 1 21 19"/>'),
    list: svg('<path d="M9 6h11M9 12h11M9 18h11"/><circle cx="4.5" cy="6" r="1"/><circle cx="4.5" cy="12" r="1"/><circle cx="4.5" cy="18" r="1"/>'),
    lock: svg('<rect x="5" y="11" width="14" height="9" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/>'),
    plus: svg('<path d="M12 5v14M5 12h14"/>'),
    x: svg('<path d="M6 6l12 12M18 6L6 18"/>'),
    user: svg('<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>'),
    phone: svg('<path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.7a2 2 0 0 1-.5 2.1L8 9.8a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.7.7a2 2 0 0 1 1.7 2z"/>'),
    cake: svg('<path d="M4 21h16v-7a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2z"/><path d="M4 16c2 1.3 3.3 1.3 5.3 0s3.3-1.3 5.4 0 3.3 1.3 5.3 0"/><path d="M12 12V8M12 5.5v-.01"/>'),
    idcard: svg('<rect x="3" y="5" width="18" height="14" rx="2"/><circle cx="9" cy="11" r="2"/><path d="M6 16a3 3 0 0 1 6 0M15 10h3M15 13h3"/>'),
    megaphone: svg('<path d="M3 11v2a1 1 0 0 0 1 1h2l5 4V6L6 10H4a1 1 0 0 0-1 1z"/><path d="M15 8.5a5 5 0 0 1 0 7"/><path d="M18 6a8.5 8.5 0 0 1 0 12"/>'),
    coordinator: svg('<circle cx="9" cy="8" r="3.5"/><path d="M3 20a6 6 0 0 1 12 0M16 11l2 2 4-4"/>'),
    adult: svg('<circle cx="12" cy="4.5" r="2.5"/><path d="M12 8v7M8 11h8M12 15l-3 6M12 15l3 6"/>'),
    child: svg('<circle cx="12" cy="8" r="2.5"/><path d="M12 11v5M9 13h6M12 16l-2 4M12 16l2 4"/>'),
    baby: svg('<circle cx="12" cy="12" r="8"/><path d="M9.5 11h.01M14.5 11h.01M9.5 14.5a3.5 3.5 0 0 0 5 0"/>'),
    family: svg('<circle cx="7" cy="6" r="2.5"/><circle cx="17" cy="6" r="2.5"/><circle cx="12" cy="12" r="2"/><path d="M3 21v-5a4 4 0 0 1 8 0"/><path d="M13 21v-5a4 4 0 0 1 8 0"/><path d="M9.5 21v-2.5a2.5 2.5 0 0 1 5 0V21"/>'),
    elder: svg('<circle cx="10" cy="4.5" r="2.2"/><path d="M10 7.5l-1 6 3 3v5M9 13.5l-3 7.5M10 9l4 2M16 11v10"/>'),
    disabled: svg('<circle cx="11" cy="4" r="2"/><path d="M11 7v6h5l3 6M11 10h5"/><path d="M8 11.5a5 5 0 1 0 6.5 6.5"/>'),
    drop: svg('<path d="M12 3s6 7 6 11a6 6 0 0 1-12 0c0-4 6-11 6-11z"/>'),
    food: svg('<path d="M7 3v8M5 3v5a2 2 0 0 0 4 0V3M7 11v10M17 3c-2 2-2 6 0 8v10"/>'),
    tent: svg('<path d="M3 20L12 4l9 16H3zM9 20l3-6 3 6"/>'),
    repair: svg('<path d="M2 12l6.5-5.5L15 12"/><path d="M4 10.5V20h9v-9.5"/><path d="M7 20v-3.5h3V20"/><path d="M16 5h6v3h-6z"/><path d="M19 8v13"/>'),
    pot: svg('<path d="M4 10h16v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-6zM2 10h2M20 10h2M9 6c0-1 1-1 1-2M14 6c0-1 1-1 1-2"/>'),
    bottle: svg('<path d="M10 2h4M10 5h4l1 3v12a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2V8l1-3zM9 12h6"/>'),
    fuel: svg('<path d="M4 21V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v16M3 21h12M14 9h2a2 2 0 0 1 2 2v6a1.5 1.5 0 0 0 3 0V8l-3-3M7 7h4"/>'),
    locate: svg('<circle cx="12" cy="12" r="4"/><circle cx="12" cy="12" r="8"/><path d="M12 1v3M12 20v3M1 12h3M20 12h3"/>'),
    other: svg('<circle cx="12" cy="12" r="9"/><path d="M12 8v8M8 12h8"/>'),
    gas: svg('<path d="M9 3h6v3H9z"/><path d="M7 6h10a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2z"/><path d="M12 10c1.5 1.6 2 2.7 2 3.7a2 2 0 0 1-4 0c0-1 .5-2.1 2-3.7z"/>'),
    glass: svg('<path d="M6 3h12l-1.5 17a1 1 0 0 1-1 1h-7a1 1 0 0 1-1-1z"/><path d="M6.5 9h11"/>'),
    meal: svg('<path d="M3 13h18a9 9 0 0 1-18 0z"/><path d="M8 3c-1 1.5 1 2.5 0 4M12 3c-1 1.5 1 2.5 0 4M16 3c-1 1.5 1 2.5 0 4"/>'),
    pan: svg('<circle cx="10" cy="13" r="6"/><path d="M16 13h6"/><path d="M8 11c.5-.6 1.5-1 2.5-.8"/>'),
    edit: svg('<path d="M4 20h4L19 9l-4-4L4 16z"/><path d="m13.5 6.5 4 4"/>'),
    sim: svg('<path d="M7 3h7l5 5v12a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/><path d="M9 11h6v6H9z"/><path d="M12 11v6M9 14h6"/>'),
    // Lista de empaque.
    box: svg('<path d="M3 7l9-4 9 4v10l-9 4-9-4V7z"/><path d="M3 7l9 4 9-4M12 11v10"/>'),
    foodbag: svg('<path d="M5 8h14l-1 13H6L5 8z"/><path d="M9 8V6a3 3 0 0 1 6 0v2"/><path d="M9 13h6M10 16.5h4"/>'),
    hygiene: svg('<rect x="3" y="11" width="10" height="10" rx="2"/><path d="M6 11V8h4v3M8 8V5h3"/><path d="M18 21V10M16 3h4v7h-4z"/>'),
    cleaning: svg('<path d="M15 2l-3 9"/><path d="M7 11h9l2 10H5l2-10z"/><path d="M9 15v6M13 15v6"/>'),
    flashlight: svg('<path d="M7 3h10v4l-2 3v11H9V10L7 7V3z"/><path d="M7 7h10M12 13v3"/>'),
    repellent: svg('<rect x="6" y="9" width="9" height="12" rx="2"/><path d="M8.5 9V6h4v3M12.5 6h2.5"/><path d="M18 4h.01M20.5 6h.01M18 8h.01"/>'),
    chair: svg('<path d="M8 3v9h8V3"/><path d="M6 12h12M7 12l-1 9M17 12l1 9M7.5 16h9"/>'),
    hammock: svg('<path d="M3 4v17M21 4v17"/><path d="M3 8c3 7 15 7 18 0"/><path d="M3 8l3 4M21 8l-3 4"/>'),
    net: svg('<path d="M12 3L4 20h16L12 3z"/><path d="M12 3v17M8.5 20L12 3l3.5 17M7.3 13.5h9.4"/>'),
    plug: svg('<path d="M9 2v5M15 2v5"/><path d="M6 7h12v4a6 6 0 0 1-12 0V7z"/><path d="M12 17v5"/>'),
    camping: svg('<path d="M12 3c2.5 3 4 4.5 4 7.5a4 4 0 0 1-8 0c0-2 1-3 2-4 0 1.8.8 2.8 2 3 0-2-1-4 0-6.5z"/><path d="M4 21l16-4M20 21L4 17"/>'),
    towel: svg('<path d="M4 4h16v4H4z"/><path d="M6 8v13h12V8"/><path d="M6 15h12"/>'),
    poncho: svg('<path d="M12 3a3 3 0 0 0-3 3v1L3 18h18L15 7V6a3 3 0 0 0-3-3z"/><path d="M9 7h6M12 18v3"/>'),
    umbrella: svg('<path d="M3 12a9 9 0 0 1 18 0H3z"/><path d="M12 12v7a2 2 0 0 0 4 0M12 3V2"/>'),
    medicine: svg('<rect x="3" y="7" width="18" height="13" rx="2"/><path d="M9 7V4h6v3M12 10.5v6M9 13.5h6"/>'),
    utensils: svg('<path d="M6 3v6a2 2 0 0 0 4 0V3M8 3v18"/><path d="M17 3c-2 0-3 2-3 5s1 4 3 4 3-1 3-4-1-5-3-5zM17 12v9"/>'),
    disposable: svg('<path d="M6 4h12l-2 17H8L6 4z"/><path d="M5 4h14M6.6 9h10.8"/>'),
    mattress: svg('<rect x="2" y="8" width="20" height="9" rx="3"/><path d="M5 17v2.5M19 17v2.5M6.5 12.5h.01M10 12.5h.01M14 12.5h.01M17.5 12.5h.01"/>')
  };
  const style = `.pf-thumb{position:relative;margin:0;width:76px}.pf-thumb a{display:grid;place-items:center;width:76px;height:76px;border-radius:9px;overflow:hidden;background:#e7eef0;border:1px solid #d5e0e2;text-decoration:none}.pf-thumb img{width:100%;height:100%;object-fit:cover}.pf-thumb-icon{font-size:26px;color:#173c46}
.pfp{border:0;border-radius:14px;padding:0;width:min(460px,calc(100% - 24px));max-height:calc(92dvh / var(--text-zoom,1));color:#173c46;font:inherit}@media(max-width:760px){.pfp{inset:0;width:auto;height:auto;max-width:none;max-height:none;margin:0;border-radius:0}.pfp-box{height:100%;box-sizing:border-box;padding:max(14px,env(safe-area-inset-top)) 14px 14px}.pfp-pick{display:flex;flex-direction:column;min-height:0;flex:1 1 auto}.pfp-pick:has(+.pfp-qty:not([hidden])),.pfp-pick[hidden]{flex:none}.pfp .pfp-options{max-height:none;flex:1 1 auto;min-height:140px}.pfp-actions{margin-top:auto;padding-top:10px;background:#fff}}.pfp::backdrop{background:#10263ab8}.pfp-box{padding:20px;display:flex;flex-direction:column;gap:6px}.pfp h2{margin:0;font-size:20px}#pfPackTitle,.pfp-pick>label{position:absolute;width:1px;height:1px;margin:-1px;overflow:hidden;clip:rect(0 0 0 0);white-space:nowrap}.pfp-for{display:none}.pfp-pick input{margin-top:0}.pfp h3{margin:6px 0 0;font-size:16px}.pfp-for,.pfp-hint{margin:0;color:#667982;font-size:13px}
.pfp label{display:block;font-size:12px;font-weight:750;margin:8px 0 0}.pfp input{width:100%;padding:11px;border:1px solid #cad9d9;border-radius:9px;font:inherit;background:#fff;box-sizing:border-box}
.pfp-options{max-height:min(300px,calc(40dvh / var(--text-zoom,1)));overflow:auto;margin-top:6px;border:1px solid #e1e9ea;border-radius:10px}.pfp-options button{display:flex;align-items:center;gap:9px;width:100%;margin:0;padding:10px 12px;border:0;border-radius:0;border-bottom:1px solid #eef3f3;background:#fff;color:#173c46;font:inherit;font-weight:650;text-align:left;cursor:pointer}.pfp-options button:last-child{border-bottom:0}.pfp-options button.hl,.pfp-options button:hover{background:#eef6f4}.pfp-options svg,.pfp-line svg,.pfp-chosen svg{flex:none;color:#2d7f86}.pfp-options .pfp-new{color:#19745f}
.pfp-chosen{display:flex;align-items:center;gap:8px;font-size:16px;margin-top:4px}.pfp-qtyrow{display:flex;gap:6px}.pfp-qtyrow input{text-align:center;font-size:18px;font-weight:800}.pfp-step{flex:none;width:48px;margin:0;border:1px solid #cad9d9;border-radius:9px;background:#fff;font-size:22px;font-weight:800;color:#173c46;cursor:pointer}
.pfp-list{display:flex;flex-direction:column;gap:4px;margin-top:8px}.pfp-line{display:flex;align-items:center;gap:8px;padding:7px 9px;border:1px solid #e1e9ea;border-radius:9px;background:#fbfdfd}.pfp-line span{flex:1;font-weight:650;overflow-wrap:anywhere}.pfp-line input{width:86px;padding:6px;text-align:center}.pfp-del{margin:0;border:0;background:none;color:#9a4536;font-size:15px;padding:4px 6px;cursor:pointer}
.pfp-btn{border:0;border-radius:10px;background:#19745f;color:#fff;font:inherit;font-weight:800;padding:12px 16px;margin:8px 0 0;width:100%;cursor:pointer}.pfp-btn.pfp-plain{background:#eef3f3;color:#173c46}.pfp-btn:disabled{opacity:.5}.pfp-actions{display:flex;gap:8px;flex-wrap:wrap;margin-top:8px}.pfp-actions .pfp-btn{flex:1 1 120px;width:auto;margin:0}.pfp-msg{margin:0;color:#a33527;font-size:13px}
.pf-pack{margin:8px 0}.pf-pack-title{display:flex;align-items:center;gap:6px;font-size:12px;color:#173c46;margin-bottom:5px}.pf-pack-title svg{width:15px;height:15px;color:#2d7f86}.pf-pack-chips{display:flex;flex-wrap:wrap;gap:5px}.pf-pack-chip{display:inline-flex;align-items:center;gap:5px;border:1px solid #c9dde0;border-radius:16px;padding:3px 9px;font-size:12px;font-weight:650;color:#1f3a46;background:#f3f9f9}.pf-pack-chip svg{width:15px;height:15px;color:#2d7f86;flex:none}.pf-pack-chip b{color:#0f6f63}
.pf-gpslive{display:flex;align-items:center;gap:8px;margin:8px 0 0;padding:8px 10px;border-radius:10px;background:#f1f6f5;border:1px solid #dbe7e5;font-size:13px;font-weight:650;color:#36505a}.pf-gl-dot{flex:none;width:9px;height:9px;border-radius:50%;background:#9aa9ad;animation:pfglpulse 1.2s ease-in-out infinite}.pf-gl-t{flex:1;min-width:0}.pf-gl-mini{flex:none;width:56px;height:6px;border-radius:4px;background:#e1e9e9;overflow:hidden}.pf-gl-mini i{display:block;height:100%;width:0;background:#9aa9ad;transition:width .4s}.pf-gpslive[data-q=good]{background:#eaf6f0;border-color:#bfe2cf}.pf-gpslive[data-q=good] .pf-gl-dot,.pf-gpslive[data-q=good] i{background:#23834f;animation:none}.pf-gpslive[data-q=mid] .pf-gl-dot,.pf-gpslive[data-q=mid] i{background:#d69a1b}.pf-gpslive[data-q=low] .pf-gl-dot,.pf-gpslive[data-q=low] i{background:#c62828}.pf-gpslive[data-q=off] .pf-gl-dot{animation:none}@keyframes pfglpulse{50%{opacity:.35}}.pf-section{border:1px solid #e1e9ea;border-radius:12px;padding:10px 14px 14px;margin:0 0 12px;background:#fbfdfd;min-width:0}
.pf-section legend{display:flex;align-items:center;gap:7px;font-weight:800;font-size:15px;color:#173c46;padding:0 6px}
.pf-section legend svg{color:#2d7f86}.pf-step{display:inline-grid;place-items:center;width:22px;height:22px;border-radius:50%;background:#1d6f7a;color:#fff;font-size:12px;font-weight:800}.pf-admin{background:#fffaf4;border-color:#ecd9c8}.pf-admin legend svg{color:#b5653f}
.pf-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(118px,1fr));gap:10px}.pf-grid .wide{grid-column:1/-1}
.pf label{display:block;margin:0;font-size:12px;font-weight:700;color:#36505a}.pf .pf-ico{display:flex;align-items:center;gap:5px}.pf .pf-ico svg{color:#2d7f86;flex:none}
.pf input,.pf textarea,.pf select{display:block;width:100%;max-width:none;min-height:0;margin-top:4px;padding:9px 10px;border:1px solid #d5e0e2;border-radius:8px;background:#fff;font:inherit;font-weight:400;color:#152b3a}
.pf textarea{min-height:64px;resize:vertical}.pf input:focus,.pf textarea:focus,.pf select:focus{outline:2px solid #2d7f8655;border-color:#2d7f86}
.pf-hint{font-size:12px;color:#5b6f76;margin:2px 0 10px;line-height:1.4}.pf-hint b{color:#9a4c2f}.pf-sub{font-size:12px;font-weight:800;color:#2d7f86;text-transform:uppercase;letter-spacing:.05em;margin:14px 0 8px}
.pf .pf-gps{display:flex;align-items:center;justify-content:center;gap:8px;width:100%;margin:2px 0 0;padding:13px;border:0;border-radius:10px;background:#1d6f7a;color:#fff;font-size:15px;font-weight:800}
.pf .pf-gps:hover{background:#175d66}.pf .pf-gps:disabled{opacity:.6;cursor:wait}.pf-gpsthumb{height:140px;margin:8px 0 0;border-radius:10px;overflow:hidden;border:1px solid #cad9d9;cursor:pointer;background:#dfe8e8}.pf-gpsthumb[hidden]{display:none}.pf-gpsthumb .leaflet-control-attribution{font-size:9px}.pf-thumbpin{width:24px;height:24px;background:#c62828;border:3px solid #fff;border-radius:50% 50% 50% 0;transform:rotate(-45deg);box-shadow:0 2px 7px #0006}.pf-gpsmsg{font-size:12px;color:#36505a;margin:6px 0 0}.pf-gpsmsg:empty{display:none}
.pf-me-btns{display:inline-flex;gap:4px;margin-left:auto}.pf .pf-ico:has(.pf-me-btns){width:100%}.pf .pf-yo{margin:0;padding:2px 9px;border:1px solid #2d7f86;border-radius:12px;background:#eaf6f3;color:#1d6f7a;font:inherit;font-size:11px;font-weight:800;cursor:pointer}.pf .pf-clearcoord{display:inline-flex;margin:0;padding:3px;border:0;border-radius:6px;background:none;color:#8a9ca1;cursor:pointer}.pf .pf-clearcoord svg{width:14px;height:14px}.pf .pf-clearcoord:hover{background:#f2e7e3;color:#9a4536}.pf-members{margin-top:8px}.pf label[hidden]{display:none!important}.pf-coordme{display:flex;align-items:center;gap:12px;padding:12px 14px;border:1.5px solid #b9d3d8;border-radius:12px;background:#f3faf9}.pf-cm-ico{display:grid;place-items:center;width:36px;height:36px;border-radius:50%;background:#e3f4ec;color:#19745f;flex:none}.pf-cm-txt{flex:1;min-width:0;display:flex;flex-direction:column;line-height:1.3}.pf-cm-txt small{font-size:11px;font-weight:800;text-transform:uppercase;letter-spacing:.04em;color:#64808a}.pf-cm-txt b{font-size:16px;color:#173c46}.pf-cm-sub{font-size:13px;color:#4b6670}.pf .pf-cm-change{flex:none;margin:0;padding:8px 12px;border:1px solid #cad9d9;border-radius:9px;background:#fff;color:#173c46;font-weight:800;font-size:13px;cursor:pointer}.pf .pf-backme{margin-left:8px;padding:2px 8px;border:1px solid #b9d3d8;border-radius:999px;background:#fff;color:#19745f;font-size:12px;font-weight:800;cursor:pointer}.pf-member{display:grid;grid-template-columns:22px 2fr 1.4fr .7fr 1.3fr 26px;grid-template-areas:'num name phone age ced del';gap:6px;align-items:center;padding:4px 0}.pf-mnum{grid-area:num;font-size:13px;font-weight:800;color:#2d7f86;text-align:right}.pf-m-name{grid-area:name}.pf-m-phone{grid-area:phone}.pf-m-age{grid-area:age}.pf-m-cedula{grid-area:ced}.pf-member .pf-remove{grid-area:del}.pf-member:last-child .pf-remove{visibility:hidden}.pf .pf-member input{min-width:0;margin-top:0}.pf-mf{position:relative;display:block;min-width:0}.pf .pf-hold,.pf-hold{display:flex;align-items:flex-start;gap:12px;margin:14px 0 4px;padding:12px 14px;border:1.5px solid #e9c46a;border-radius:12px;background:#fff8e1;cursor:pointer}.pf-hold input{position:absolute;opacity:0;width:1px;height:1px}.pf-hold-sw{display:block;position:relative;flex:none;width:40px;height:24px;margin-top:2px;border-radius:999px;background:#cad9d9;transition:background .15s}.pf-hold-sw:after{content:"";position:absolute;top:3px;left:3px;width:18px;height:18px;border-radius:50%;background:#fff;box-shadow:0 1px 3px #0003;transition:left .15s}.pf-hold input:checked+.pf-hold-sw{background:#a87400}.pf-hold input:checked+.pf-hold-sw:after{left:19px}.pf-hold input:focus-visible+.pf-hold-sw{outline:2px solid #1f63c6;outline-offset:2px}.pf-hold-txt{display:flex;flex-direction:column;gap:2px;font-size:14px;color:#4a3a10}.pf-hold-txt small{font-size:12px;color:#7a5600}.pf .pf-addmember{display:inline-flex;align-items:center;gap:6px;margin:8px 0 0 28px;padding:8px 12px;border:1px dashed #9fb6bb;border-radius:9px;background:#f7fafa;color:#19745f;font:inherit;font-size:13px;font-weight:800;cursor:pointer}.pf .pf-mf input{width:100%;padding-left:32px}.pf-mi{position:absolute;left:9px;top:50%;transform:translateY(-50%);display:flex;color:#7a9298;pointer-events:none}.pf-mi svg{width:16px;height:16px}
.pf .pf-remove{display:inline-flex;margin:0;padding:3px;border:0;border-radius:6px;background:none;color:#8a9ca1}.pf .pf-remove:hover{background:#f2e7e3;color:#9a4536}
.pf-needs{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px}.pf-needgroup{grid-column:1/-1;margin:6px 0 -2px;font-size:11px;font-weight:800;letter-spacing:.05em;text-transform:uppercase;color:#64808a}.pf-needgroup:first-child{margin-top:0}.pf-needitem{min-width:0}.pf-needitem.has-opts:has(input[data-need]:checked){grid-column:1/-1;border:1px solid #2d7f86;border-radius:11px;background:#f6fbfa}.pf-needitem.has-opts:has(input[data-need]:checked) .pf-need span{border:0;background:none}.pf-needitem .pf-datanet{margin:0;padding:0 10px 10px}
.pf-datanet-label small{font-weight:600;color:#667982;margin-left:4px}.pf-pills{display:flex;flex-wrap:wrap;gap:6px}.pf-pill{display:inline-flex;flex-direction:row;align-items:center;gap:7px;min-height:42px;padding:8px 10px;border:1.5px solid #cad9d9;border-radius:10px;background:#fff;color:#173c46;text-align:left;cursor:pointer;font:inherit}.pf-pill b{font-size:14px}.pf-pill small{font-size:11px;color:#667982;font-weight:500}.pf-pill[aria-pressed="true"]{border-color:#19745f;background:#e3f4ec}.pf-pill svg{color:#2d7f86;flex:none}.pf-swatch{display:inline-block;width:12px;height:12px;border-radius:50%;background:#b7c7ca;flex:none}.pf-swatch.pf-net-mat{display:none}.pf-other-input[hidden]{display:none}.pf-pill-other{border-style:dashed}.pf-other-input{display:block;width:100%;max-width:260px;margin:8px 0 0;padding:8px 10px;border:1px solid #cad9d9;border-radius:8px;font-size:15px;background:#fff}.pf-pill[aria-pressed="true"] b:before{content:"✓ ";color:#19745f}
.pf label.pf-need{position:relative;margin:0;font-weight:650}.pf .pf-need input{position:absolute;opacity:0;width:1px;height:1px;margin:0}
.pf-need span{display:flex;align-items:center;gap:8px;min-height:44px;padding:8px 10px;border:1px solid #d5e0e2;border-radius:10px;background:#fff;color:#36505a;font-size:13px;cursor:pointer}
.pf-need span svg{color:#2d7f86;flex:none}.pf-need span:after{content:"";margin-left:auto;width:16px;height:16px;border:1.5px solid #b7c7ca;border-radius:4px;flex:none}
.pf-need input:checked+span{border-color:#2d7f86;background:#eaf6f3;color:#173c46}.pf-need input:checked+span:after{background:#2d7f86 url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='3'%3E%3Cpath d='M5 12l5 5 9-10'/%3E%3C/svg%3E") center/12px no-repeat;border-color:#2d7f86}
.pf-chips{display:flex;flex-wrap:wrap;gap:5px;margin:6px 0}.pf-chip{display:inline-flex;align-items:center;gap:5px;border:1px solid #d5e0e2;border-radius:16px;padding:3px 9px;font-size:13px;font-weight:650;color:#1f3a46;background:#fff}.pf-chip svg{color:#2d7f86;flex:none}.pf-chip.text{border-style:dashed}
.pf-verified{display:inline-flex;align-items:center;gap:4px;vertical-align:middle;margin-left:5px;line-height:1}.pf-verified svg{flex:none;filter:drop-shadow(0 1px 1px #a8740055)}.pf-verified.tap{border:0;background:none;padding:2px;margin-left:3px;cursor:pointer;border-radius:50%}.pf-verified.tap:hover,.pf-verified.tap[aria-expanded="true"]{background:#fff3cd}
.pf-vpop{position:absolute;z-index:4000;max-width:260px;background:#fff;border:1px solid #e9c46a;border-radius:10px;box-shadow:0 10px 26px #17354233;padding:8px 10px;font:700 13px/1.35 Inter,system-ui,sans-serif;color:#5a4100}
.pf-netbadge{display:inline-flex;align-items:center;margin-left:6px;padding:1px 7px;border-radius:999px;background:#e7eef0;color:#36505a;font:800 11px/1.5 system-ui,sans-serif;letter-spacing:.01em;white-space:nowrap;text-transform:none}.pf-net-tigo{background:#00377b;color:#fff;box-shadow:inset 0 -2px 0 #00c8ff}.pf-net-masmovil{background:#0153ba;color:#fff}.pf-net-beber{background:#d8ecfb;color:#134d7a}.pf-net-cocinar{background:#d9f2ea;color:#14594a}.pf-net-preparada{background:#fde3cf;color:#8a3a09}.pf-net-ingredientes{background:#e6f3d5;color:#3c5a14}.pf-net-gasolina{background:#f2a900;color:#2b1d00}.pf-net-diesel{background:#2f3a40;color:#fff}.pf-net-tropigas{background:#e35205;color:#fff}.pf-net-panagas{background:#272974;color:#fff;box-shadow:inset 0 -2px 0 #e11e26}
.pf-datanet{margin-top:10px}.pf-datanet-label{display:flex;align-items:center;gap:6px;font-size:13px;font-weight:750;color:#173c46;margin-bottom:5px}.pf-datanet-label svg{color:#2d7f86}
.pf-need input:focus-visible+span{outline:2px solid #2d7f86}.pf-repair span svg{color:#1f63c6}.pf-repair input:checked+span{border-color:#1f63c6;background:#e9f1fc}.pf-repair input:checked+span:after{background-color:#1f63c6;border-color:#1f63c6}.pf-need em{font-style:normal}.pf-need em small{display:block;font-size:11px;font-weight:700;color:#ae4f32}.pf-need em small.pf-dup{color:#677c83}.pf-otros[hidden]{display:none}.pf label.pf-otros{margin-top:8px}
.pf .pf-num{display:flex;margin-top:4px}.pf .pf-num input{flex:1;width:auto;min-width:0;margin:0;border-radius:0;text-align:center}
.pf .pf-numbtn{flex:none;width:36px;margin:0;padding:0;border:1px solid #d5e0e2;background:#f1f6f6;color:#1d6f7a;font-size:18px;font-weight:800;line-height:1;cursor:pointer}.pf .pf-numbtn:hover{background:#e2efee}
.pf .pf-dec{order:-1;border-radius:8px 0 0 8px;border-right:0}.pf .pf-inc{border-radius:0 8px 8px 0;border-left:0}
dialog.pf-mapdialog{inset:0;width:auto;height:auto;max-width:none;max-height:none;margin:0;padding:0;border:0;background:#dfe8e8}dialog.pf-mapdialog::backdrop{background:#10263ab8}
.pf-mapdialog .pf-mapbox{width:100%;height:100%}.pf-mapdialog .pf-mapclose{position:absolute;top:12px;right:12px;z-index:1000;display:flex;margin:0;padding:9px;border:0;border-radius:50%;background:#fff;color:#173c46;box-shadow:0 2px 8px #0004;cursor:pointer}
.pf-mapdialog .pf-mapgps{position:absolute;top:60px;right:12px;z-index:1000;display:flex;align-items:center;gap:6px;margin:0;padding:9px 12px;border:0;border-radius:20px;background:#fff;color:#1f6fd1;font:inherit;font-size:13px;font-weight:800;box-shadow:0 2px 8px #0004;cursor:pointer}.pf-mapdialog .pf-mapgps:disabled{opacity:.6;cursor:wait}
.pf-mapdialog .pf-mappanel{position:absolute;left:12px;right:12px;bottom:22px;z-index:1000;margin:0 auto;max-width:520px;padding:10px 12px;border-radius:12px;background:#fffffff2;color:#173c46;font-size:13px;box-shadow:0 2px 10px #0004}
.pf-mapdialog .pf-mapstatus,.pf-mapdialog .pf-mappick{margin:0 0 6px}.pf-mapdialog .pf-mappick{font-weight:700}
.pf-mapdialog .pf-mapuse{display:block;width:100%;margin:0;padding:11px;border:0;border-radius:9px;background:#1d6f7a;color:#fff;font:inherit;font-size:15px;font-weight:800;cursor:pointer}.pf-mapdialog .pf-mapuse:disabled{opacity:.5;cursor:not-allowed}
.pf-mapdialog .pf-crosshair{position:absolute;top:50%;left:50%;z-index:1000;width:44px;height:44px;transform:translate(-50%,-50%);pointer-events:none;color:#c62828;filter:drop-shadow(0 0 1.5px #fff) drop-shadow(0 0 1.5px #fff)}
.pf-me{width:22px;height:22px;border-radius:50%;background:#1f6fd1;border:3px solid #fff;box-shadow:0 0 0 2px #1f6fd1,0 1px 6px #0006}
@media(max-width:560px){.pf-section{padding:8px 10px 12px}.pf-grid{gap:8px;grid-template-columns:minmax(0,1fr)}.pf-grid:has([data-filter="count"]){grid-template-columns:repeat(2,minmax(0,1fr))}.pf-member{grid-template-columns:22px minmax(0,1fr) minmax(0,1.6fr) 26px;grid-template-areas:'num name name del' '. phone phone .' '. age ced .';row-gap:4px;padding:6px 0;border-top:1px dashed #dbe5e6}.pf-member:first-child{border-top:0}.pf-needs{grid-template-columns:repeat(auto-fill,minmax(9.5em,1fr))}}`;
  if (!document.getElementById('pointform-style')) document.head.insertAdjacentHTML('beforeend', `<style id="pointform-style">${style}</style>`);

  // filter: «count» y «age» solo dígitos (0–999); «phone» dígitos, espacios, +, - y paréntesis.
  const input = (name, label, opt = {}) => `<label class="${opt.wide ? 'wide' : ''}"><span class="${opt.icon ? 'pf-ico' : ''}">${opt.icon ? icons[opt.icon] : ''}${esc(label)}</span>${opt.area
    ? `<textarea name="${name}" placeholder="${esc(opt.ph || '')}"></textarea>`
    : opt.filter === 'count' ? countInput(name, label, opt.ph)
    : `<input name="${name}"${filterAttrs(opt.filter)} placeholder="${esc(opt.ph || '')}" autocomplete="off">`}</label>`;
  // Conteo con botones − / +. El campo va primero en el HTML para que el <label> lo enlace a él y no al botón; el CSS pone «−» a la izquierda.
  const countInput = (name, label, ph) => `<span class="pf-num"><input name="${name}"${filterAttrs('count')} placeholder="${esc(ph || '')}" autocomplete="off"><button type="button" class="pf-numbtn pf-dec" data-step="-1" aria-label="Restar 1 a ${esc(label)}">−</button><button type="button" class="pf-numbtn pf-inc" data-step="1" aria-label="Sumar 1 a ${esc(label)}">+</button></span>`;
  const filterAttrs = f => f === 'count' || f === 'age' ? ` data-filter="${f}" inputmode="numeric" maxlength="3"` : f === 'phone' ? ' data-filter="phone" type="tel" inputmode="tel" maxlength="40"' : '';
  // Los tres pasos llevan número (1, 2, 3); la sección de administración lleva candado.
  const section = (icon, title, body, cls = '') => `<fieldset class="pf-section ${cls}"><legend>${typeof icon === 'number' ? `<span class="pf-step">${icon}</span>` : icons[icon]}${esc(title)}</legend>${body}</fieldset>`;
  // «Reparación de casa» es solo para administración: la atienden constructores y no se publica.
  const NEEDS = [['Agua', 'drop'], ['Alimentos', 'food'], ['Refugio (carpa / lona / toldo)', 'tent'], ['Kit de cocina', 'pot'], ['Kit de bebé', 'bottle'], ['Combustible para generador', 'fuel'], ['Tarjeta de datos prepago', 'sim'], ['Gas de cocina', 'gas'], ['Reparación de casa', 'repair', 'pf-repair']];
  // Necesidades con marca: la red o la empresa va entre paréntesis, p. ej. «Tarjeta de datos prepago (Tigo)» o
  // «Gas de cocina (Panagas)». Se elige de la lista o se escribe otra. Insignia con el color de cada marca.
  // Redes móviles activas en Panamá: Tigo y +Móvil (Claro pasó a +Móvil en 2023; Digicel salió en 2024).
  // Grupos del formulario (dos columnas; al marcar una necesidad con opciones, ocupa toda la fila y las abre debajo).
  const NEED_GROUPS = [['Agua y comida', ['Agua', 'Alimentos', 'Kit de cocina', 'Gas de cocina']], ['Refugio y cuidado', ['Refugio (carpa / lona / toldo)', 'Kit de bebé']], ['Energía y comunicación', ['Combustible para generador', 'Tarjeta de datos prepago']], ['Casa', ['Reparación de casa']]];
  // Opciones de una necesidad: multi (se marcan una o varias: Agua para beber o para cocinar) o una sola con lista y
  // «Otra…» para escribir (combustible, red, empresa del gas).
  // [nombre, clase de color, texto de la insignia, alias, ícono]. other: botón «Otra…» con un campo corto para escribir.
  const BRANDED = {
    'Agua': {key: 'water', label: '¿Para qué?', brands: [['Para beber', 'beber', 'para beber', null, 'glass'], ['Para cocinar', 'cocinar', 'para cocinar', null, 'pot']]},
    'Alimentos': {key: 'food', label: '¿Qué tipo?', brands: [['Comida preparada', 'preparada', 'comida preparada', null, 'meal'], ['Ingredientes para cocinar', 'ingredientes', 'ingredientes', null, 'pan']]},
    'Combustible para generador': {key: 'fuel', label: 'Combustible', other: 'Otro', otherPh: 'Escribe cuál', def: 'Gasolina', brands: [['Gasolina', 'gasolina', 'Gasolina'], ['Diésel', 'diesel', 'Diésel']]},
    'Tarjeta de datos prepago': {key: 'net', label: 'Red', other: 'Otra', otherPh: 'Escribe la red', brands: [['Tigo', 'tigo', 'tigo'], ['+Móvil', 'masmovil', '+móvil', /^(mas|\+) ?movil$|^claro$/]]},
    'Gas de cocina': {key: 'gas', label: 'Empresa', other: 'Otra', otherPh: 'Escribe la empresa', brands: [['Tropigas', 'tropigas', 'Tropigas'], ['Panagas', 'panagas', 'Panagas']]},
    'Reparación de casa': {key: 'repair', label: 'Materiales (opcional)', other: 'Otro', otherPh: 'Escribe el material', brands: [['Hojas de zinc', 'mat', 'hojas de zinc'], ['Carriola', 'mat', 'carriola'], ['Varilla', 'mat', 'varilla'], ['Tubo de hierro', 'mat', 'tubo de hierro'], ['Bolsas de cemento', 'mat', 'bolsas de cemento'], ['Bloques', 'mat', 'bloques'], ['Tornillos', 'mat', 'tornillos'], ['Clavos', 'mat', 'clavos'], ['Soldadura', 'mat', 'soldadura']]}
  };
  const DATA_NEED = 'Tarjeta de datos prepago';
  const brandedRe = new RegExp(`^\\s*(${Object.keys(BRANDED).map(l => l.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})\\s*(?:\\(([^)]*)\\))?\\s*$`, 'i');
  // {label, brand} si la parte es una necesidad con marca (brand puede ser ''); si no, null.
  // Antes eran dos casillas: «Gasolina para generador» y «Diésel para generador»; se leen como combustible con su tipo.
  const LEGACY = [[/^\s*gasolina para generador\s*$/i, 'Combustible para generador', 'Gasolina'], [/^\s*di[eé]sel para generador\s*$/i, 'Combustible para generador', 'Diésel']];
  const brandOf = part => { const old = LEGACY.find(([re]) => re.test(String(part || ''))); if (old) return {label: old[1], brand: old[2]}; const m = brandedRe.exec(String(part || '')); if (!m) return null; const label = Object.keys(BRANDED).find(l => l.toLowerCase() === m[1].toLowerCase()); return {label, brand: (m[2] || '').trim()}; };
  const findBrand = (label, b) => (BRANDED[label]?.brands || []).find(x => fold(x[0]) === fold(b) || fold(x[2]) === fold(b) || (x[3] && x[3].test(fold(b))));
  const brandBadge = (label, b) => { if (!b) return ''; if (String(b).includes(' / ')) return String(b).split(' / ').map(v => brandBadge(label, v.trim())).join(''); const x = findBrand(label, b); return x ? `<span class="pf-netbadge pf-net-${x[1]}" title="${esc(x[0])}">${esc(x[2])}</span>` : `<span class="pf-netbadge" title="${esc(b)}">${esc(b)}</span>`; };
  const dataNet = part => { const x = brandOf(part); return x && x.label === DATA_NEED ? x.brand : null; };
  const netBadge = b => brandBadge(DATA_NEED, b);
  const needBase = part => brandOf(part)?.label || String(part || '').trim();
  const needIconOf = label => icons[(NEEDS.find(n => n[0] === label) || [, 'other'])[1]];
  // Chip de una necesidad con marca (para las listas de cada página); '' si no lo es.
  const needChip = (part, cls = 'chip') => { const x = brandOf(part); return x ? `<span class="${cls}">${needIconOf(x.label)}${esc(x.label)}${brandBadge(x.label, x.brand)}</span>` : ''; };
  // «Verificado por equipo»: insignia dorada cuando alguien de un equipo creó o actualizó el punto o la necesidad (revisión
  // más a fondo). Quien es de un equipo ve qué equipo y quién (el servidor solo se lo manda a ellos); los demás y el
  // público, solo el ícono («Verificado», sin equipo, persona ni fecha).
  const AWARD = '<svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="9" r="6.5" fill="#f2b705" stroke="#a87400" stroke-width="1.5"/><path d="M8.2 14.2 6.5 22l5.5-3 5.5 3-1.7-7.8" fill="#f2b705" stroke="#a87400" stroke-width="1.5" stroke-linejoin="round"/><path d="m9.3 9 1.9 1.9 3.6-3.7" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>';
  const verifiedTitle = x => { const d = new Date(x.verifiedAt), when = isNaN(d) ? '' : d.toLocaleDateString('es-PA', {day: 'numeric', month: 'short'}); return x.verifiedTeam ? `Verificado por el equipo ${x.verifiedTeam.replace(/^equipo\s+/i, '')}${x.verifiedBy ? ' · ' + x.verifiedBy : ''}${when ? ' · ' + when : ''}` : 'Verificado'; };
  // Siempre solo el ícono. Con los datos del equipo (solo llegan a quien es de un equipo) es un botón: tocarlo muestra
  // qué equipo, quién y cuándo. Sin ellos (público, sin equipo) es solo el ícono, sin detalles.
  const verifiedBadge = x => !x || !x.verifiedAt ? '' : x.verifiedTeam
    ? `<button type="button" class="pf-verified tap" data-verified="${esc(verifiedTitle(x))}" aria-expanded="false" aria-label="Verificado: ver qué equipo">${AWARD}</button>`
    : `<span class="pf-verified" title="Verificado" role="img" aria-label="Verificado">${AWARD}</span>`;
  const closeVerified = () => { document.querySelectorAll('.pf-vpop').forEach(x => x.remove()); document.querySelectorAll('.pf-verified.tap[aria-expanded="true"]').forEach(b => b.setAttribute('aria-expanded', 'false')); };
  // En captura: el toque en la insignia no abre ni cierra la fila o la tarjeta donde está.
  if (typeof document.addEventListener === 'function') document.addEventListener('click', e => {
    const b = e.target.closest('.pf-verified.tap');
    if (!b) { if (!e.target.closest('.pf-vpop')) closeVerified(); return; }
    e.preventDefault(); e.stopPropagation();
    const open = b.getAttribute('aria-expanded') === 'true';
    closeVerified();
    if (open) return;
    const pop = document.createElement('div');
    pop.className = 'pf-vpop';
    pop.setAttribute('role', 'status');
    pop.textContent = b.dataset.verified;
    document.body.appendChild(pop);
    b.setAttribute('aria-expanded', 'true');
    const r = b.getBoundingClientRect();
    pop.style.top = (r.bottom + scrollY + 6) + 'px';
    pop.style.left = Math.max(8, Math.min(r.left + scrollX, document.documentElement.clientWidth - pop.offsetWidth - 8)) + 'px';
  }, true);
  if (typeof document.addEventListener === 'function') document.addEventListener('scroll', closeVerified, true);
  // Fichas de una lista de necesidades (texto «Agua (Para beber), Arroz»): ícono y opción de cada tipo; lo demás como
  // texto. Para páginas sin su propia versión (ficha pública).
  const needChipsHtml = text => `<div class="pf-chips">${String(text || '').split(/[,\n]+/).map(p => p.trim()).filter(Boolean).map(part => needChip(part, 'pf-chip') || (hit => hit ? `<span class="pf-chip">${icons[hit[1]]}${esc(hit[0])}</span>` : `<span class="pf-chip text">${esc(part)}</span>`)(NEEDS.find(([l]) => fold(l) === fold(part)))).join('')}</div>`;
  // Servicios del punto que se repiten: nombre, ícono y cada cuántos días se sugiere.
  const SERVICES = [['Recolección de basura', 'trash', 7], ['Cambio de baño portátil', 'toilet', 7], ['Fumigación de mosquitos', 'spray', 14], ['Ronda médica', 'medicine', 30]];
  const serviceIcon = name => icons[(SERVICES.find(s => s[0] === name) || [, 'other'])[1]] || '';
  // Estado de un servicio para mostrarlo igual en todas las páginas: cuenta regresiva hasta la próxima vez.
  function serviceStatus(n, now = Date.now()) {
    if (n.status === 'IN_PROGRESS') return {cls: 'route', text: 'En camino'};
    if (n.status === 'COMPLETED') return {cls: 'ok', text: 'Completado'};
    const due = Date.parse(n.nextDueAt || '');
    if (!due || due <= now) { const late = due ? Math.floor((now - due) / 864e5) : 0; return late >= 1 ? {cls: 'late', text: `Atrasado ${late} ${late === 1 ? 'día' : 'días'}`} : {cls: 'due', text: 'Toca hoy'}; }
    const days = Math.ceil((due - now) / 864e5);
    return days <= 1 ? {cls: 'soon', text: 'Mañana'} : {cls: days <= 2 ? 'soon' : 'ok', text: `En ${days} días`};
  }
  const serviceScheduled = (n, now = Date.now()) => n.kind === 'service' && n.status === 'OPEN' && Date.parse(n.nextDueAt || '') > now;
  const COUNTS = [['people', 'Personas', 'people'], ['adults', 'Adultos', 'adult'], ['children', 'Niños', 'child'], ['babies', 'Bebés', 'baby'], ['olderAdults', 'Mayores', 'elder'], ['disabled', 'Discapacitados', 'disabled'], ['families', 'Familias', 'family']];

  // Casillas de necesidades con iconos y «Otros»; también las usa el formulario de necesidades del panel.
  const pillMark = x => x[4] && icons[x[4]] ? icons[x[4]] : `<i class="pf-swatch pf-net-${esc(x[1])}"></i>`;
  const optsHtml = l => { const b = BRANDED[l]; if (!b) return ''; return `<div class="pf-datanet" data-brandfor="${esc(l)}" hidden><span class="pf-datanet-label">${esc(b.label)} <small>toca una o varias</small></span><div class="pf-pills" role="group" aria-label="${esc(l)}: ${esc(b.label)}">${b.brands.map(x => `<button type="button" class="pf-pill" data-pill="${esc(x[0])}" aria-pressed="false">${pillMark(x)}<b>${esc(x[0])}</b></button>`).join('')}${b.other ? `<button type="button" class="pf-pill pf-pill-other" data-pill-other aria-pressed="false">${icons.edit || '✏️'}<b>${esc(b.other)}</b></button>` : ''}</div>${b.other ? `<input class="pf-other-input" maxlength="40" placeholder="${esc(b.otherPh)}" aria-label="${esc(l)}: ${esc(b.otherPh)}" hidden>` : ''}<input type="hidden" class="pf-net-input" data-brand="${esc(l)}"></div>`; };
  const needItem = ([l, i, c]) => `<div class="pf-needitem${BRANDED[l] ? ' has-opts' : ''}"><label class="pf-need ${c || ''}"><input type="checkbox" data-need="${esc(l)}"><span>${icons[i]}${esc(l)}</span></label>${optsHtml(l)}</div>`;
  // only: solo esas necesidades (p. ej. al editar una reparación: la casilla fija y sus materiales).
  const needsHtml = ({admin = false, only = null} = {}) => { const shown = NEEDS.filter(n => admin || n[3] !== 'admin'); if (only) return `<div class="pf-needs">${shown.filter(n => only.includes(n[0])).map(needItem).join('')}<input type="checkbox" class="pf-otros-check" hidden></div><label class="pf-otros" hidden><textarea class="pf-otros-text"></textarea></label>`; return `<div class="pf-needs">${NEED_GROUPS.map(([g, ls]) => [g, shown.filter(n => ls.includes(n[0]))]).concat([['', shown.filter(n => !NEED_GROUPS.some(([, ls]) => ls.includes(n[0])))]]).filter(([, list]) => list.length).map(([g, list]) => `${g ? `<p class="pf-needgroup">${esc(g)}</p>` : ''}${list.map(needItem).join('')}`).join('')}<p class="pf-needgroup">¿Algo más?</p><div class="pf-needitem"><label class="pf-need"><input type="checkbox" class="pf-otros-check"><span>${icons.other}Otros</span></label></div></div><label class="pf-otros" hidden>Otros: ¿qué más necesitan?<textarea class="pf-otros-text" placeholder="Medicinas, ropa, colchones…"></textarea></label>`; };
  function html({admin = false, edit = false} = {}) {
    const where = section(1, '¿Dónde estás?', `<button type="button" class="pf-gps">${icons.pin}Enviar mi ubicación GPS</button><div class="pf-gpsthumb" hidden title="Toca para mover la ubicación" aria-label="Vista previa de la ubicación elegida"></div><p class="pf-gpsmsg" role="status"></p><div class="pf-grid" style="margin-top:12px">${input('gps', 'o pega un enlace de Google Maps', {wide: true, ph: 'https://maps.google.com/…'})}${input('area', 'Zona / área', {ph: 'Ej. Cañas'})}${input('place', 'Ubicación / referencia', {ph: 'Ej. Escuela de la comunidad'})}${input('directions', 'Indicaciones de llegada', {wide: true, area: true, ph: 'Cómo llegar, puntos de referencia…'})}</div>`);
    const who = section(2, '¿Con quién estás?', `<div class="pf-grid">${input('coordinator', 'Reportado por', {icon: 'megaphone', ph: 'Quien más sabe de este punto'})}${input('coordinatorPhone', 'Teléfono de quien reporta', {filter: 'phone', ph: 'Ej. 6000-0000'})}</div><p class="pf-sub">¿Cuántas personas?</p><div class="pf-grid">${COUNTS.map(([k, l, i]) => input(k, l, {icon: i, filter: 'count', ph: '0'})).join('')}</div><p class="pf-hint"><b>Al menos un contacto con teléfono es obligatorio.</b></p><div class="pf-members"></div><button type="button" class="pf-addmember">＋ Agregar otro contacto</button>`);
    const what = section(3, '¿Qué necesitas?', `${needsHtml({admin})}<div class="pf-grid" style="margin-top:12px">${input('extra', 'Notas adicionales', {wide: true, area: true})}</div>`);
    const adminOnly = admin ? section('lock', 'Solo administración', `<div class="pf-grid">${input('code', 'Código', {ph: 'Se asigna uno si lo dejas vacío'})}${input('row', 'N.º fila')}${input('ordinal', 'Orden')}<label>Servicio reportado<select name="service"><option value="">Sin confirmar</option><option value="YES">Sí (reportado)</option><option value="MAYBE">Tal vez</option><option value="NO">No</option></select></label>${input('serviceNotes', 'Notas de servicio', {wide: true, area: true})}${input('notes', 'Notas de grupo', {wide: true, area: true})}${edit ? input('note', 'Nota privada del campamento (opcional)', {wide: true, area: true}) : ''}</div>`, 'pf-admin') : '';
    return where + who + what + adminOnly;
  }
  const memberFields = [['name', 'Nombre', ''], ['phone', 'Teléfono', 'phone'], ['age', 'Edad', 'age'], ['cedula', 'Cédula', '']];
  // Ícono de cada campo de la persona (nombre, teléfono, edad, cédula), dentro del campo a la izquierda.
  const memberIcons = {name: 'user', phone: 'phone', age: 'cake', cedula: 'idcard'};
  // Una fila por persona: número a la izquierda y los campos con su nombre como texto de ayuda (sin título «Persona N»).
  const memberRow = (m, i) => `<div class="pf-member"><span class="pf-mnum" aria-hidden="true">${i + 1}</span>${memberFields.map(([k, l, f]) => `<span class="pf-mf pf-m-${k}"><span class="pf-mi" aria-hidden="true">${icons[memberIcons[k]]}</span><input data-k="${k}" value="${esc(m[k])}" placeholder="${l}" aria-label="${l} de la persona ${i + 1}" autocomplete="off"${filterAttrs(f)}></span>`).join('')}<button type="button" class="pf-remove" aria-label="Quitar persona ${i + 1}" title="Quitar">${icons.x}</button></div>`;
  const filters = {count: v => v.replace(/\D/g, '').slice(0, 3), age: v => v.replace(/\D/g, '').slice(0, 3), phone: v => v.replace(/[^\d+\-\s()]/g, '')};
  // Puntos anteriores guardaban hasta tres contactos sueltos; se muestran como personas para editarlos.
  function membersOf(p = {}) {
    if (Array.isArray(p.members) && p.members.length) return p.members;
    return [[p.contact, p.phone], [p.secondContact, p.secondPhone], [p.thirdContact, p.thirdPhone]].filter(([n, t]) => n || t).map(([name, phone]) => ({name: name || '', phone: phone || ''}));
  }
  // «Necesidades» se guarda como texto: opciones marcadas y lo escrito en «Otros», separados por comas.
  // Lista de empaque: artículos iniciales con su ícono (los mismos nombres que defaultPackingItems en packing.go).
  const PACKING = [['Bolsas alimentación', 'foodbag'], ['Kit aseo personal', 'hygiene'], ['Kit limpieza', 'cleaning'], ['Kit bebé', 'bottle'], ['Linternas', 'flashlight'], ['Carpas / toldas', 'tent'], ['Repelente', 'repellent'], ['Sillas', 'chair'], ['Hamacas', 'hammock'], ['Mosquiteros', 'net'], ['Extensiones', 'plug'], ['Camping', 'camping'], ['Toallas', 'towel'], ['Capotes / ponchos', 'poncho'], ['Paraguas', 'umbrella'], ['Medicamentos', 'medicine'], ['Utensilios', 'utensils'], ['Desechables', 'disposable'], ['Colchonetas', 'mattress']];
  const packIcon = name => { const hit = PACKING.find(([l]) => fold(l) === fold(String(name || ''))); return icons[hit ? hit[1] : 'box']; };
  // packingTotals suma las cantidades por artículo (sin distinguir mayúsculas ni acentos), en el orden en que se agregaron.
  const packingTotals = lines => { const out = []; (lines || []).forEach(l => { const hit = out.find(x => fold(x.item) === fold(l.item)); if (hit) hit.quantity += l.quantity; else out.push({item: l.item, quantity: l.quantity}); }); return out; };
  const packingHtml = lines => { const totals = packingTotals(lines); return totals.length ? `<div class="pf-pack"><b class="pf-pack-title">${icons.box}Lista de empaque</b><div class="pf-pack-chips">${totals.map(t => `<span class="pf-pack-chip">${packIcon(t.item)}${esc(t.item)} <b>× ${t.quantity}</b></span>`).join('')}</div></div>` : ''; };
  // Diálogo «Agregar a lista de empaque» (corredores, agentes y administración):
  // elegir un artículo o escribir uno nuevo → cantidad → repetir → revisar → guardar.
  // openPacking({subtitle, items, save}) — save(lines) guarda y devuelve una promesa; si falla, el diálogo muestra el error.
  const pk = {items: [], draft: [], pick: '', save: null};
  const pq = sel => document.querySelector('#pfPack ' + sel);
  const packDraftHtml = edit => pk.draft.length ? `<div class="pfp-list">${pk.draft.map((l, i) => `<div class="pfp-line">${packIcon(l.item)}<span>${esc(l.item)}</span>${edit ? `<input type="number" min="1" max="100000" step="1" inputmode="numeric" value="${l.quantity}" data-qty="${i}" aria-label="Cantidad de ${esc(l.item)}">` : `<b>× ${l.quantity}</b>`}<button type="button" class="pfp-del" data-del="${i}" aria-label="Quitar ${esc(l.item)}">✕</button></div>`).join('')}</div>` : '<p class="pfp-hint">Todavía no agregaste artículos.</p>';
  function packDraw(edit) { const n = pk.draft.length; pq('[data-act=review]').disabled = !n; pq('[data-act=review]').textContent = n ? `Revisar (${n})` : 'Revisar'; pq('.pfp-draft').innerHTML = packDraftHtml(edit); }
  function packOptions() {
    const raw = pq('#pfPackSearch').value.trim().replace(/\s+/g, ' '), q = fold(raw), list = pk.items.filter(i => !q || fold(i.name).includes(q)), exact = pk.items.some(i => fold(i.name) === q);
    pq('.pfp-options').innerHTML = list.map(i => `<button type="button" role="option" data-item="${esc(i.name)}">${packIcon(i.name)}${esc(i.name)}</button>`).join('') + (raw && !exact ? `<button type="button" role="option" class="pfp-new" data-item="${esc(raw)}">${icons.box}Agregar «${esc(raw)}» como artículo nuevo</button>` : '') + (!list.length && !raw ? '<p class="pfp-hint">No hay artículos guardados todavía.</p>' : '');
    pq('.pfp-options [data-item]')?.classList.add('hl');
  }
  function packStep(step) {
    pq('.pfp-pick').hidden = step !== 'pick'; pq('.pfp-qty').hidden = step !== 'qty'; pq('.pfp-review').hidden = step !== 'review';
    pq('[data-act=review]').hidden = step === 'review'; pq('[data-act=save]').hidden = pq('[data-act=more]').hidden = step !== 'review';
    pq('.pfp-msg').textContent = ''; packDraw(step === 'review');
    if (step === 'pick') { pq('#pfPackSearch').value = ''; packOptions(); pq('#pfPackSearch').focus(); }
  }
  function packChoose(name) {
    const known = pk.items.find(i => fold(i.name) === fold(name)); pk.pick = known ? known.name : name;
    pq('.pfp-chosen').innerHTML = `${packIcon(pk.pick)}<b>${esc(pk.pick)}</b>${known ? '' : ' <small>(nuevo)</small>'}`;
    pq('#pfPackQty').value = 1; packStep('qty'); pq('#pfPackQty').focus(); pq('#pfPackQty').select();
  }
  function packAdd() {
    const qty = Number(pq('#pfPackQty').value);
    if (!Number.isInteger(qty) || qty < 1 || qty > 100000) { pq('.pfp-msg').textContent = 'Escribe una cantidad entera mayor que cero.'; pq('#pfPackQty').focus(); return; }
    const hit = pk.draft.find(l => fold(l.item) === fold(pk.pick)); if (hit) hit.quantity += qty; else pk.draft.push({item: pk.pick, quantity: qty});
    packStep('pick');
  }
  const packDiscard = () => !pk.draft.length || confirm('¿Descartar los artículos sin guardar?');
  async function packSave() {
    const bad = pk.draft.find(l => !Number.isInteger(l.quantity) || l.quantity < 1 || l.quantity > 100000);
    if (bad) { pq('.pfp-msg').textContent = `Revisa la cantidad de «${bad.item}»: debe ser un número entero mayor que cero.`; return; }
    const b = pq('[data-act=save]'); b.disabled = true;
    try { await pk.save(pk.draft.map(l => ({item: l.item, quantity: l.quantity}))); pk.draft = []; document.getElementById('pfPack').close(); }
    catch (err) { pq('.pfp-msg').textContent = 'No se pudo guardar: ' + err.message; }
    finally { b.disabled = false; }
  }
  function packDialog() {
    let d = document.getElementById('pfPack');
    if (d) return d;
    document.body.insertAdjacentHTML('beforeend', `<dialog id="pfPack" class="pfp" aria-labelledby="pfPackTitle"><div class="pfp-box"><h2 id="pfPackTitle">Agregar a lista de empaque</h2><p class="pfp-for"></p><div class="pfp-pick"><label for="pfPackSearch">Artículo</label><input id="pfPackSearch" autocomplete="off" spellcheck="false" placeholder="Escribe para buscar o agregar uno nuevo…" role="combobox" aria-controls="pfPackOptions" aria-expanded="true"><div id="pfPackOptions" class="pfp-options" role="listbox" aria-label="Artículos"></div></div><div class="pfp-qty" hidden><div class="pfp-chosen"></div><label for="pfPackQty">Cantidad</label><div class="pfp-qtyrow"><button type="button" class="pfp-step" data-step="-1" aria-label="Menos">−</button><input id="pfPackQty" type="number" min="1" max="100000" step="1" inputmode="numeric" value="1"><button type="button" class="pfp-step" data-step="1" aria-label="Más">+</button></div><button type="button" class="pfp-btn" data-act="add">Agregar a la lista</button><button type="button" class="pfp-btn pfp-plain" data-act="other">Elegir otro artículo</button></div><div class="pfp-review" hidden><h3>Revisa antes de guardar</h3><p class="pfp-hint">Ajusta cantidades o quita artículos. Al guardar se suman a la lista de empaque de esta necesidad.</p></div><div class="pfp-draft"></div><p class="pfp-msg" role="status"></p><div class="pfp-actions"><button type="button" class="pfp-btn pfp-plain" data-act="cancel">Cancelar</button><button type="button" class="pfp-btn pfp-plain" data-act="more" hidden>Seguir agregando</button><button type="button" class="pfp-btn" data-act="review" disabled>Revisar</button><button type="button" class="pfp-btn" data-act="save" hidden>Guardar lista</button></div></div></dialog>`);
    d = document.getElementById('pfPack');
    pq('#pfPackSearch').addEventListener('input', packOptions);
    pq('#pfPackSearch').addEventListener('keydown', e => {
      const opts = [...d.querySelectorAll('.pfp-options [data-item]')], at = opts.findIndex(o => o.classList.contains('hl'));
      if (e.key === 'Enter') { e.preventDefault(); if (opts[at]) packChoose(opts[at].dataset.item); }
      else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); if (!opts.length) return; opts[at]?.classList.remove('hl'); const next = opts[(at + (e.key === 'ArrowDown' ? 1 : -1) + opts.length) % opts.length]; next.classList.add('hl'); next.scrollIntoView({block: 'nearest'}); }
    });
    pq('.pfp-options').addEventListener('click', e => { const b = e.target.closest('[data-item]'); if (b) packChoose(b.dataset.item); });
    pq('.pfp-qty').addEventListener('click', e => { const b = e.target.closest('[data-step]'); if (b) { const q = pq('#pfPackQty'); q.value = Math.max(1, (parseInt(q.value, 10) || 0) + Number(b.dataset.step)); } });
    pq('#pfPackQty').addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); packAdd(); } });
    pq('.pfp-draft').addEventListener('click', e => { const b = e.target.closest('[data-del]'); if (!b) return; pk.draft.splice(Number(b.dataset.del), 1); const reviewing = !pq('.pfp-review').hidden; if (reviewing && !pk.draft.length) packStep('pick'); else packDraw(reviewing); });
    pq('.pfp-draft').addEventListener('input', e => { const i = e.target.dataset.qty; if (i !== undefined) pk.draft[Number(i)].quantity = Number(e.target.value); });
    d.addEventListener('click', e => {
      const act = e.target.closest('[data-act]')?.dataset.act;
      if (act === 'add') packAdd(); else if (act === 'other' || act === 'more') packStep('pick');
      else if (act === 'review' && pk.draft.length) packStep('review'); else if (act === 'save') packSave();
      else if (act === 'cancel' && packDiscard()) d.close();
    });
    d.addEventListener('cancel', e => { if (!packDiscard()) e.preventDefault(); });
    return d;
  }
  function openPacking({subtitle = '', items = [], save}) {
    const d = packDialog(); pk.items = items; pk.draft = []; pk.save = save;
    pq('.pfp-for').textContent = subtitle; d.showModal(); packStep('pick');
  }
  // Fotos y videos de una carrera (partida / entrega). Las fotos grandes se reducen a 2000 px en JPEG antes de subir
  // para gastar menos datos; los videos y lo que el navegador no pueda abrir (p. ej. HEIC) se suben tal cual.
  async function prepareMedia(file) {
    if (!/^image\/(jpeg|png|webp)$/.test(file.type) || file.size < 1.5e6 || typeof createImageBitmap !== 'function') return file;
    try {
      const img = await createImageBitmap(file), scale = Math.min(1, 2000 / Math.max(img.width, img.height));
      const canvas = document.createElement('canvas'); canvas.width = Math.round(img.width * scale); canvas.height = Math.round(img.height * scale);
      canvas.getContext('2d').drawImage(img, 0, 0, canvas.width, canvas.height);
      const blob = await new Promise(r => canvas.toBlob(r, 'image/jpeg', 0.85));
      return blob && blob.size < file.size ? new File([blob], file.name.replace(/\.\w+$/, '') + '.jpg', {type: 'image/jpeg'}) : file;
    } catch { return file; }
  }
  // uploadMedia sube el archivo tal cual (sin multipart) y avisa el avance (0–1); devuelve el registro guardado.
  function uploadMedia(url, file, {csrf = '', onProgress = () => {}} = {}) {
    return new Promise((resolve, reject) => {
      const x = new XMLHttpRequest();
      x.open('POST', url); x.withCredentials = true;
      x.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
      x.setRequestHeader('X-File-Name', encodeURIComponent(file.name || ''));
      if (csrf) x.setRequestHeader('X-CSRF-Token', csrf);
      x.upload.onprogress = e => { if (e.lengthComputable) onProgress(e.loaded / e.total); };
      x.onload = () => { let body = {}; try { body = JSON.parse(x.responseText); } catch {} if (x.status >= 200 && x.status < 300) resolve(body); else reject(Object.assign(new Error(body.error || `Error HTTP ${x.status}`), {status: x.status})); };
      x.onerror = () => reject(new Error('Sin conexión. Intenta de nuevo cuando tengas señal.'));
      x.send(file);
    });
  }
  // mediaThumbs dibuja las miniaturas de una etapa; src(item) da la URL privada (panel de corredor o de administración).
  const mediaThumbs = (items, src, extra = () => '') => items.map(m => `<figure class="pf-thumb"><a href="${esc(src(m))}" target="_blank" rel="noopener noreferrer" title="${esc(m.name || '')}">${String(m.type).startsWith('image/') ? `<img src="${esc(src(m))}" alt="" loading="lazy" onerror="this.replaceWith(Object.assign(document.createElement('span'),{className:'pf-thumb-icon',textContent:'📷'}))">` : '<span class="pf-thumb-icon">▶</span>'}</a>${extra(m)}</figure>`).join('');
  const fold = s => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().trim();
  function parseNeeds(text) {
    const checked = new Set(), other = [], brands = {};
    for (const part of String(text || '').split(',').map(s => s.trim()).filter(Boolean)) {
      const hit = NEEDS.find(([l]) => fold(l) === fold(part)), br = brandOf(part);
      if (br) { checked.add(br.label); if (br.brand) brands[br.label] = [brands[br.label], br.brand].filter(Boolean).join(' / '); } else if (hit) checked.add(hit[0]); else other.push(part);
    }
    return {checked: [...checked], other: other.join(', '), brands, net: brands[DATA_NEED] || ''};
  }
  const needsText = (checked, other) => [...checked, String(other || '').trim()].filter(Boolean).join(', ');
  // Marca las casillas según el texto guardado. Una opción sin casilla aquí (p. ej. «Reparación de casa» fuera de
  // administración) pasa a «Otros» para que guardar no la borre.
  function fillNeeds(root, text) {
    const needs = parseNeeds(text), boxes = [...root.querySelectorAll('[data-need]')], shown = boxes.map(b => b.dataset.need);
    const otros = root.querySelector('.pf-otros'), otrosCheck = root.querySelector('.pf-otros-check'), otrosText = root.querySelector('.pf-otros-text');
    for (const box of boxes) box.checked = needs.checked.includes(box.dataset.need);
    otrosText.value = needsText(needs.checked.filter(l => !shown.includes(l)), needs.other);
    // «Otros» marcado ⇔ hay texto: escribir lo marca y desmarcarlo borra el texto, así nunca se pierde en silencio.
    const sync = () => { otrosCheck.checked = !!otrosText.value.trim(); otros.hidden = !otrosCheck.checked; };
    sync();
    otrosCheck.onchange = () => { otros.hidden = !otrosCheck.checked; if (otrosCheck.checked) otrosText.focus(); else otrosText.value = ''; };
    otrosText.oninput = () => { if (otrosText.value.trim()) sync(); };
    root.querySelectorAll('.pf-datanet').forEach(box => wireBrand(root, box, needs.brands[box.dataset.brandfor]));
  }
  // Opciones de una necesidad: botones que se marcan (una o varias); «Otra» abre un campo corto para escribir. El valor
  // («Tigo / Digicel») va en un campo oculto que lee readNeeds. Combustible marca Gasolina al marcar la casilla.
  function wireBrand(root, box, value) {
    const label = box.dataset?.brandfor, b = BRANDED[label]; if (!b) return;
    const check = root.querySelector(`[data-need="${label}"]`); if (!check) return;
    const input = box.querySelector('.pf-net-input'), pills = [...box.querySelectorAll('[data-pill]')], otherPill = box.querySelector('[data-pill-other]'), otherIn = box.querySelector('.pf-other-input');
    const on = (el, v) => el && el.setAttribute('aria-pressed', String(v)), isOn = el => el?.getAttribute('aria-pressed') === 'true';
    const compose = () => { const other = otherIn && isOn(otherPill) ? otherIn.value.replace(/[(),/]/g, ' ').replace(/\s+/g, ' ').trim() : ''; input.value = [...pills.filter(isOn).map(p => p.dataset.pill), other].filter(Boolean).join(' / '); };
    const parts = String(value || '').split('/').map(v => v.trim()).filter(Boolean), known = parts.map(v => findBrand(label, v)?.[0]).filter(Boolean), unknown = parts.filter(v => !findBrand(label, v));
    pills.forEach(p => on(p, known.includes(p.dataset.pill)));
    if (otherPill) { on(otherPill, unknown.length > 0); otherIn.value = unknown.join(' '); otherIn.hidden = !unknown.length; }
    compose();
    const sync = () => { box.hidden = !check.checked; };
    sync();
    if (box.dataset.wired) return; box.dataset.wired = '1';
    check.addEventListener('change', () => { sync(); if (check.checked && b.def && !input.value) { on(pills.find(p => p.dataset.pill === b.def), true); compose(); } });
    box.addEventListener('click', e => { const p = e.target.closest('[data-pill],[data-pill-other]'); if (!p) return; on(p, !isOn(p)); if (p === otherPill) { otherIn.hidden = !isOn(p); if (isOn(p)) otherIn.focus(); } compose(); });
    otherIn?.addEventListener('input', compose);
  }
  function readNeeds(root) {
    const brandIn = l => { const v = (root.querySelector(`.pf-net-input[data-brand="${l}"]`)?.value || '').replace(/[(),]/g, ' ').replace(/\s+/g, ' ').trim(), known = v && findBrand(l, v); return known ? known[0] : v || BRANDED[l]?.def || ''; };
    const checked = [...root.querySelectorAll('[data-need]:checked')].map(b => b.dataset.need).map(l => BRANDED[l] && brandIn(l) ? `${l} (${brandIn(l)})` : l);
    const otrosCheck = root.querySelector('.pf-otros-check'), otrosText = root.querySelector('.pf-otros-text');
    if (otrosText.value.trim()) otrosCheck.checked = true;
    return needsText(checked, otrosText.value);
  }
  function readMembers(root) {
    return [...root.querySelectorAll('.pf-member')].map(row => Object.fromEntries(memberFields.map(([k]) => [k, row.querySelector(`[data-k="${k}"]`).value.trim()])));
  }
  // Siempre queda una fila vacía al final: al escribir en ella aparece la siguiente, lista para llenar.
  const memberFilled = m => memberFields.some(([k]) => String(m[k] || '').trim());
  function drawMembers(root, list) {
    list = list.filter(memberFilled);
    list.push({});
    root.querySelector('.pf-members').innerHTML = list.map(memberRow).join('');
  }
  // Distancia en metros entre dos [lat, lng].
  function meters(a, b) {
    const r = Math.PI / 180, dLat = (b[0] - a[0]) * r, dLng = (b[1] - a[1]) * r;
    const h = Math.sin(dLat / 2) ** 2 + Math.cos(a[0] * r) * Math.cos(b[0] * r) * Math.sin(dLng / 2) ** 2;
    return 2 * 6371000 * Math.asin(Math.sqrt(h));
  }
  // Puntos con coordenadas ordenados del más cercano al más lejano, con su distancia.
  function nearest(here, list) {
    return (list || []).filter(p => Array.isArray(p.coords) && p.coords.length === 2).map(p => ({p, d: meters(here, p.coords)})).sort((a, b) => a.d - b.d);
  }
  const distance = d => d < 1000 ? `${Math.round(d)} m` : `${(d / 1000).toFixed(1)} km`;
  const PIN_COLORS = {approved: '#277a66', pending: '#e57450'};
  const inPanama = ([lat, lng]) => lat >= 6 && lat <= 10 && lng >= -83 && lng <= -78;
  // Coordenadas de un enlace de mapa ya escrito en el formulario («…?q=7.45,-80.35»), o null.
  function linkCoords(text) {
    const m = String(text || '').match(/(-?\d{1,2}\.\d+)\s*,\s*(-?\d{1,3}\.\d+)/);
    return m ? [Number(m[1]), Number(m[2])] : null;
  }
  const mapLink = ([lat, lng]) => `https://maps.google.com/?q=${lat.toFixed(6)},${lng.toFixed(6)}`;
  const crosshair = svg('<circle cx="12" cy="12" r="6"/><path d="M12 1v6M12 17v6M1 12h6M17 12h6"/>').replace('width="18" height="18"', 'width="44" height="44"');
  // Ventana de pantalla completa con solo el mapa (solo si se pasó `nearby`): los demás puntos en verde (aprobados) o naranja
  // (pendientes). Mientras llega el GPS, o si no se obtiene, queda centrado en el último punto agregado; `me` pone tu
  // ubicación en azul y centra el mapa en ella, salvo que ya hayas movido el mapa. Si el formulario ya tiene una ubicación
  // (elegida antes, GPS o enlace pegado), abre ahí y el GPS no la mueve ni la reemplaza (`keep`).
  // La cruz fija en el centro marca el punto elegido: se mueve arrastrando el mapa o tocando un lugar, y «Usar esta ubicación»
  // lo pasa al enlace GPS del formulario.
  function openGpsMap(root) {
    if (!root._nearby || !window.L) return null;
    let d = document.querySelector('dialog.pf-mapdialog');
    if (!d) {
      document.body.insertAdjacentHTML('beforeend', `<dialog class="pf-mapdialog" aria-label="Mapa de puntos y tu ubicación"><div class="pf-mapbox"></div><div class="pf-crosshair" aria-hidden="true">${crosshair}</div><button type="button" class="pf-mapclose" aria-label="Cerrar mapa" title="Cerrar">${icons.x}</button><button type="button" class="pf-mapgps" title="Mover la cruz a tu ubicación GPS">${icons.locate}Mi GPS</button><div class="pf-mappanel"><p class="pf-mapstatus" role="status"></p><p class="pf-mappick"></p><button type="button" class="pf-mapuse">Usar esta ubicación</button></div></dialog>`);
      d = document.querySelector('dialog.pf-mapdialog');
      d.querySelector('.pf-mapclose').onclick = () => d.close();
    }
    const list = (root._nearby() || []).filter(p => Array.isArray(p.coords) && p.coords.length === 2);
    const last = lastAdded(list), status = d.querySelector('.pf-mapstatus'), pick = d.querySelector('.pf-mappick'), use = d.querySelector('.pf-mapuse');
    if (d._map) d._map.remove();
    d.showModal();
    const map = d._map = L.map(d.querySelector('.pf-mapbox'));
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {maxZoom: 19, attribution: '© OpenStreetMap contributors'}).addTo(map);
    for (const p of list) L.circleMarker(p.coords, {radius: 8, weight: 2, color: '#fff', fillColor: PIN_COLORS[p.status] || PIN_COLORS.approved, fillOpacity: 1}).addTo(map).bindTooltip(`${esc(p.code)} · ${esc(p.area)}${p.status === 'pending' ? ' · pendiente' : ''}${p === last ? ' · último agregado' : ''}`);
    const prev = linkCoords(root.querySelector('[name="gps"]').value), keep = !!prev && inPanama(prev);
    let moved = keep, gpsHere = null, gpsLayers = [];
    const center = () => { const c = map.getCenter(); return [c.lat, c.lng]; };
    const showPick = () => {
      const here = center(), ok = inPanama(here), near = nearest(here, list)[0];
      use.disabled = !ok;
      pick.textContent = ok ? `Cruz: ${here[0].toFixed(5)}, ${here[1].toFixed(5)}${near ? ` · a ${distance(near.d)} de ${near.p.code}` : ''}` : 'La cruz está fuera de Panamá.';
    };
    map.on('move', showPick);
    map.on('dragstart', () => { moved = true; });
    map.on('click', e => { moved = true; map.panTo(e.latlng); });
    if (keep) map.setView(prev, 17);
    else map.setView(last ? last.coords : [7.46, -80.33], last ? 15 : 11);
    setTimeout(() => map.invalidateSize(), 50);
    status.textContent = keep ? 'La cruz está en la ubicación elegida antes. Mueve el mapa o toca un lugar para cambiarla.' : 'Buscando tu ubicación… Mueve el mapa o toca un lugar para poner la cruz en el punto.';
    use.onclick = () => {
      const here = center();
      if (!inPanama(here)) return;
      root.querySelector('[name="gps"]').value = mapLink(here);
      root.querySelector('.pf-gpsmsg').textContent = '';
      d.close();
      drawGpsThumb(root);
    };
    const view = {
      keep,
      me(here, accuracy) {
        gpsHere = here;
        gpsLayers.forEach(l => map.removeLayer(l));
        gpsLayers = [L.marker(here, {icon: L.divIcon({className: '', html: '<div class="pf-me"></div>', iconSize: [22, 22]}), zIndexOffset: 1000, interactive: false}).addTo(map)];
        if (accuracy) gpsLayers.push(L.circle(here, {radius: accuracy, weight: 1, color: '#1f6fd1', fillColor: '#1f6fd1', fillOpacity: .12, interactive: false}).addTo(map));
        if (!moved) map.setView(here, 16);
        status.textContent = `Azul: tu GPS (±${Math.round(accuracy || 0)} m) · verde: aprobados · naranja: pendientes. ${keep ? 'La cruz sigue en la ubicación elegida antes.' : 'Mueve el mapa para ajustar la cruz.'}`;
      },
      status(text) { status.textContent = text + (last && !moved && !keep ? ` Mapa centrado en el último punto agregado: ${last.code} (${last.area}).` : '') + ' Puedes elegir el punto con la cruz.'; }
    };
    // «Mi GPS» lleva la cruz a tu ubicación; si aún no hay GPS lo pide. No cambia el formulario hasta «Usar esta ubicación».
    const gpsBtn = d.querySelector('.pf-mapgps');
    gpsBtn.disabled = false;
    gpsBtn.onclick = () => {
      moved = true;
      if (gpsHere) { map.setView(gpsHere, Math.max(map.getZoom(), 16)); return; }
      if (!navigator.geolocation) { status.textContent = 'Este dispositivo no permite obtener la ubicación.'; return; }
      gpsBtn.disabled = true; status.textContent = 'Buscando tu ubicación…';
      navigator.geolocation.getCurrentPosition(pos => {
        const here = [pos.coords.latitude, pos.coords.longitude];
        gpsBtn.disabled = false;
        if (!inPanama(here)) { status.textContent = 'Tu GPS queda fuera de Panamá.'; return; }
        view.me(here, pos.coords.accuracy);
        map.setView(here, Math.max(map.getZoom(), 16));
      }, err => {
        gpsBtn.disabled = false;
        status.textContent = err.code === 1 ? 'Permiso de ubicación denegado. Actívalo en el navegador.' : 'No se pudo obtener tu GPS. Intenta de nuevo.';
      }, {enableHighAccuracy: true, timeout: 20000, maximumAge: 0});
    };
    return view;
  }
  // Vista previa fija (sin arrastrar ni zoom) de la ubicación del formulario, con su pin; tocarla abre el mapa para moverla.
  // Se redibuja al elegir en el mapa, al llegar el GPS o al pegar un enlace; se ajusta sola cuando el formulario se hace visible.
  function drawGpsThumb(root) {
    const box = root.querySelector('.pf-gpsthumb');
    if (!box) return;
    const c = linkCoords(root.querySelector('[name="gps"]').value);
    if (box._map) { box._map.remove(); box._map = null; }
    box.hidden = !(c && inPanama(c) && window.L) || noMap();
    if (box.hidden) return;
    const map = box._map = L.map(box, {zoomControl: false, dragging: false, scrollWheelZoom: false, doubleClickZoom: false, boxZoom: false, keyboard: false, touchZoom: false, tap: false}).setView(c, 16);
    map.attributionControl.setPrefix(false);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {maxZoom: 19, attribution: '© OpenStreetMap'}).addTo(map);
    L.marker(c, {icon: L.divIcon({className: '', html: '<div class="pf-thumbpin"></div>', iconSize: [24, 24], iconAnchor: [12, 28]}), interactive: false}).addTo(map);
    if (window.ResizeObserver) { if (!box._watch) { box._watch = new ResizeObserver(() => { if (box._map && box.offsetWidth) { box._map.invalidateSize(); box._map.setView(box._center, 16); } }); box._watch.observe(box); } }
    box._center = c;
    setTimeout(() => { map.invalidateSize(); map.setView(c, 16); }, 50);
  }
  // Punto agregado más recientemente (por fecha de creación).
  const lastAdded = list => list.reduce((a, p) => !a || String(p.created || '') > String(a.created || '') ? p : a, null);


  // GPS en segundo plano: al abrir el formulario (¡Ayuda!, Agregar punto, punto nuevo sin conexión) el teléfono empieza a
  // buscar señal, así al tocar el botón de GPS ya hay una lectura. Se guarda la más precisa reciente; se apaga a los 5 min.
  let warm = null, warmBest = null, warmTimer = null;
  const warmSubs = new Set();
  function warmGps() {
    if (!navigator.geolocation) return;
    clearTimeout(warmTimer);
    warmTimer = setTimeout(() => { if (warm !== null) navigator.geolocation.clearWatch(warm); warm = null; }, 5 * 60000);
    if (warm !== null) return;
    warm = navigator.geolocation.watchPosition(pos => {
      if (!warmBest || pos.coords.accuracy <= warmBest.coords.accuracy || Date.now() - warmBest.timestamp > 60000) warmBest = pos;
      warmSubs.forEach(fn => fn(warmBest, null));
    }, err => { warmSubs.forEach(fn => fn(null, err)); if (err.code === 1) { navigator.geolocation.clearWatch(warm); warm = null; } }, {enableHighAccuracy: true, maximumAge: 0});
  }
  const freshFix = (ms = 120000) => warmBest && Date.now() - warmBest.timestamp < ms ? warmBest : null;
  // Sin mapa (sin conexión, o el mapa no cargó): en vez del mapa se muestra solo el GPS buscando tu posición, con la
  // precisión en vivo (verde ≤15 m, amarillo ≤50 m, rojo más). «Usar esta ubicación» toma la lectura más precisa.
  const noMap = () => navigator.onLine === false || !window.L || (typeof offOn !== 'undefined' && offOn);
  function gpsLock(root) {
    const gps = root.querySelector('[name="gps"]'), msg = root.querySelector('.pf-gpsmsg');
    if (!navigator.geolocation) { msg.textContent = 'Este dispositivo no permite obtener la ubicación. Pega un enlace de mapa.'; return; }
    if (!document.getElementById('pf-gpslock-style')) document.head.insertAdjacentHTML('beforeend', `<style id="pf-gpslock-style">.pf-gpslock{border:0;border-radius:16px;padding:0;width:min(400px,calc(100% - 24px));color:#173c46;font:15px Inter,system-ui,sans-serif;box-shadow:0 20px 60px #10263a66}.pf-gpslock::backdrop{background:#10263ad0}.pf-gl{padding:20px;display:flex;flex-direction:column;gap:12px;text-align:center}.pf-gl h2{margin:0;font-size:19px}.pf-gl p{margin:0;color:#4b6670;font-size:14px;line-height:1.45}.pf-gl-acc{font-size:44px;font-weight:900;line-height:1;color:#173c46}.pf-gl-acc small{display:block;font-size:13px;font-weight:700;color:#667982;margin-top:6px}.pf-gl-bar{height:10px;border-radius:6px;background:#e7eeee;overflow:hidden}.pf-gl-bar i{display:block;height:100%;width:0;border-radius:6px;background:#c62828;transition:width .4s,background .4s}.pf-gl-coords{font:700 13px ui-monospace,monospace;color:#36505a}.pf-gl-acts{display:flex;gap:8px}.pf-gl-acts button{flex:1;border:0;border-radius:12px;padding:13px;font:800 15px Inter,system-ui,sans-serif;cursor:pointer;background:#eef3f3;color:#173c46}.pf-gl-acts .go{background:#19745f;color:#fff}.pf-gl-acts .go:disabled{opacity:.45;cursor:default}</style>`);
    let d = document.getElementById('pfGpsLock');
    if (!d) {
      document.body.insertAdjacentHTML('beforeend', `<dialog id="pfGpsLock" class="pf-gpslock" aria-labelledby="pfGlTitle"><div class="pf-gl"><h2 id="pfGlTitle">${icons.locate} Ubicación GPS</h2><p>Sin mapa: el teléfono está buscando tu posición. Espera a que la precisión baje (menos metros es mejor) y úsala.</p><div class="pf-gl-acc" id="pfGlAcc">…<small>Buscando señal GPS</small></div><div class="pf-gl-bar"><i id="pfGlBar"></i></div><p class="pf-gl-coords" id="pfGlCoords"></p><div class="pf-gl-acts"><button type="button" data-gl="cancel">Cancelar</button><button type="button" class="go" data-gl="use" disabled>Usar esta ubicación</button></div></div></dialog>`);
      d = document.getElementById('pfGpsLock');
    }
    let best = freshFix(), n = best ? 1 : 0;
    const draw = (pos) => {
      const acc = Math.round(pos.coords.accuracy || 0), color = acc <= 15 ? '#23834f' : acc <= 50 ? '#d69a1b' : '#c62828';
      d.querySelector('#pfGlAcc').innerHTML = `±${acc} m<small>${acc <= 15 ? 'Buena precisión' : acc <= 50 ? 'Precisión media: espera un poco más' : 'Poca precisión: sal a cielo abierto si puedes'} · ${n} ${n === 1 ? 'lectura' : 'lecturas'}</small>`;
      const bar = d.querySelector('#pfGlBar'); bar.style.width = Math.max(8, Math.min(100, 100 - Math.min(acc, 200) / 2)) + '%'; bar.style.background = color;
      d.querySelector('#pfGlCoords').textContent = best ? `${best.coords.latitude.toFixed(6)}, ${best.coords.longitude.toFixed(6)}` : '';
      const use = d.querySelector('[data-gl="use"]'); use.disabled = !best; if (best) use.textContent = `Usar esta ubicación (±${Math.round(best.coords.accuracy)} m)`;
    };
    const watch = navigator.geolocation.watchPosition(pos => { n++; if (!best || pos.coords.accuracy <= best.coords.accuracy) best = pos; draw(pos); }, err => {
      d.querySelector('#pfGlAcc').innerHTML = `—<small>${err.code === 1 ? 'Permiso de ubicación denegado: actívalo en el navegador.' : 'Todavía sin señal GPS…'}</small>`;
    }, {enableHighAccuracy: true, maximumAge: 0, timeout: 30000});
    const done = () => { navigator.geolocation.clearWatch(watch); d.onclick = null; if (d.open) d.close(); };
    d.onclick = e => {
      const b = e.target.closest('[data-gl]');
      if (!b) return;
      if (b.dataset.gl === 'use' && best) {
        const c = [best.coords.latitude, best.coords.longitude];
        gps.value = mapLink(c); gps.dispatchEvent(new Event('input', {bubbles: true}));
        msg.textContent = inPanama(c) ? `✓ Ubicación GPS agregada (precisión aprox. ±${Math.round(best.coords.accuracy)} m).` : 'Tu ubicación queda fuera de Panamá; revisa el enlace antes de enviar.';
      }
      done();
    };
    d.oncancel = () => done();
    d.querySelector('[data-gl="use"]').disabled = true; d.querySelector('[data-gl="use"]').textContent = 'Usar esta ubicación';
    d.querySelector('#pfGlAcc').innerHTML = '…<small>Buscando señal GPS</small>'; d.querySelector('#pfGlBar').style.width = '0'; d.querySelector('#pfGlCoords').textContent = '';
    d.showModal();
    if (best) draw(best);
  }
  function locate(root) {
    const btn = root.querySelector('.pf-gps'), msg = root.querySelector('.pf-gpsmsg'), gps = root.querySelector('[name="gps"]');
    if (noMap()) { gpsLock(root); return; }
    const view = openGpsMap(root);
    if (!navigator.geolocation) { msg.textContent = 'Este dispositivo no permite obtener la ubicación. Pega un enlace de mapa.'; view?.status(msg.textContent); return; }
    btn.disabled = true; msg.textContent = 'Buscando tu ubicación…';
    navigator.geolocation.getCurrentPosition(pos => {
      const lat = pos.coords.latitude, lng = pos.coords.longitude, outside = !inPanama([lat, lng]);
      // Con una ubicación ya elegida en el mapa, el GPS solo se muestra; el enlace cambia con «Usar esta ubicación».
      if (view?.keep) { outside ? view.status('Tu GPS queda fuera de Panamá.') : view.me([lat, lng], pos.coords.accuracy); btn.disabled = false; return; }
      gps.value = mapLink([lat, lng]);
      drawGpsThumb(root);
      msg.textContent = outside ? 'Tu ubicación queda fuera de Panamá; revisa el enlace antes de enviar.' : `✓ Ubicación GPS agregada (precisión aprox. ±${Math.round(pos.coords.accuracy || 0)} m).`;
      if (view) outside ? view.status(msg.textContent) : view.me([lat, lng], pos.coords.accuracy);
      btn.disabled = false;
    }, err => {
      if (view?.keep) { view.status(err.code === 1 ? 'Permiso de ubicación denegado.' : 'No se pudo obtener tu GPS.'); btn.disabled = false; return; }
      msg.textContent = err.code === 1 ? 'Permiso de ubicación denegado. Actívalo en el navegador o pega un enlace de mapa.' : 'No se pudo obtener tu ubicación' + (window.isSecureContext ? '. Intenta de nuevo o pega un enlace de mapa.' : ' (requiere HTTPS).');
      view?.status(msg.textContent);
      btn.disabled = false;
    }, {enableHighAccuracy: true, timeout: 20000, maximumAge: 15000});
  }
  // Dibuja el formulario dentro de `root` (un contenedor dentro del <form>) y carga los datos de `point`.
  // `nearby`: función que devuelve los puntos del mapa de pantalla completa que se abre con «Enviar mi ubicación GPS».
  const HOLD_HOURS = 3;
  // Quien entró con código y es de un equipo queda como coordinador por defecto: se muestra como una tarjeta (sin campos)
  // con «Cambiar»; al cambiarlo aparecen los campos vacíos y «↺ Soy yo» lo devuelve. Sin equipo, los campos de siempre.
  // me = {name, telephone, team} o null.
  function setMe(root, me) {
    const coordEl = root.querySelector('[name="coordinator"]'), phoneEl = root.querySelector('[name="coordinatorPhone"]');
    if (!coordEl || !phoneEl) return;
    const grid = coordEl.closest('.pf-grid'), labels = [coordEl.closest('label'), phoneEl.closest('label')];
    root.querySelector('.pf-coordme')?.remove();
    root.querySelector('.pf-backme')?.remove();
    labels.forEach(l => { l.hidden = false; });
    // Equipo: casilla para retener la publicación unas horas (el equipo lo abastece primero); sin equipo, no aparece.
    root.querySelector('.pf-hold')?.remove();
    if (me && me.team) root.insertAdjacentHTML('beforeend', `<label class="pf-hold"><input type="checkbox" name="holdTeam"><span class="pf-hold-sw" aria-hidden="true"></span><span class="pf-hold-txt"><b>Nuestro equipo lo abastece</b><small>No se publica por ${HOLD_HOURS} horas: mientras tanto solo lo ve ${esc(me.team)}. Después se publica solo.</small></span></label>`);
    if (!me || !me.name || !me.team) return;
    const set = (name, phone) => { for (const [el, v] of [[coordEl, name], [phoneEl, phone]]) { el.value = v; el.dispatchEvent(new Event('input', {bubbles: true})); } };
    const backLink = () => { const span = labels[0].querySelector('span'); span.insertAdjacentHTML('beforeend', '<button type="button" class="pf-backme" title="Ponerme como quien reporta">↺ Soy yo</button>'); span.querySelector('.pf-backme').onclick = e => { e.preventDefault(); setMe(root, me); }; };
    // Ya tiene otro coordinador escrito (p. ej. un borrador): se respeta, con la opción de ponerse uno.
    if (coordEl.value.trim() && fold(coordEl.value) !== fold(me.name)) { backLink(); return; }
    set(me.name, me.telephone || '');
    labels.forEach(l => { l.hidden = true; });
    grid.insertAdjacentHTML('beforebegin', `<div class="pf-coordme"><span class="pf-cm-ico">${icons.megaphone}</span><div class="pf-cm-txt"><small>Reportado por</small><b>${esc(me.name)}</b><span class="pf-cm-sub">Tú · ${esc(me.team)}${me.telephone ? ' · ' + esc(me.telephone) : ''}</span></div><button type="button" class="pf-cm-change">Cambiar</button></div>`);
    root.querySelector('.pf-cm-change').onclick = e => { e.preventDefault(); root.querySelector('.pf-coordme').remove(); labels.forEach(l => { l.hidden = false; }); set('', ''); backLink(); coordEl.focus(); };
  }

  // Progreso del GPS dentro del formulario, sin bloquear: bajo el botón se ve la precisión en vivo y el campo de ubicación
  // se llena solo con la mejor lectura mientras sigues escribiendo. Si eliges un lugar en el mapa o pegas un enlace, ya
  // no se toca.
  function gpsLive(root) {
    const btn = root.querySelector('.pf-gps'), gpsEl = root.querySelector('[name="gps"]');
    if (!btn || !navigator.geolocation) return;
    let box = root.querySelector('.pf-gpslive');
    if (!box) { btn.insertAdjacentHTML('afterend', '<div class="pf-gpslive" role="status"><span class="pf-gl-dot"></span><span class="pf-gl-t">Buscando señal GPS…</span><span class="pf-gl-mini"><i></i></span></div>'); box = root.querySelector('.pf-gpslive'); }
    root._autoGps = gpsEl.value ? null : '';
    gpsEl.addEventListener('input', () => { root._autoGps = null; });
    let shown = Infinity;
    const update = (fix, err) => {
      if (!root.isConnected) { warmSubs.delete(update); return; }
      const t = box.querySelector('.pf-gl-t'), bar = box.querySelector('i');
      if (err && !fix) { if (err.code === 1) { t.textContent = 'GPS no disponible: permiso de ubicación denegado.'; box.dataset.q = 'off'; } return; }
      const acc = Math.round(fix.coords.accuracy || 0), own = root._autoGps === null && gpsEl.value;
      box.dataset.q = acc <= 15 ? 'good' : acc <= 50 ? 'mid' : 'low';
      bar.style.width = Math.max(10, Math.min(100, 100 - Math.min(acc, 200) / 2)) + '%';
      if (own) { t.textContent = `Usando la ubicación que elegiste (GPS ±${acc} m).`; return; }
      t.textContent = acc <= 15 ? `GPS ±${acc} m ✓ ubicación agregada` : `GPS ±${acc} m · mejorando…`;
      const c = [fix.coords.latitude, fix.coords.longitude];
      if (inPanama(c) && (gpsEl.value === '' || gpsEl.value === root._autoGps) && acc <= shown - 5) {
        shown = acc; gpsEl.value = root._autoGps = mapLink(c);
        if (!noMap()) drawGpsThumb(root);
      }
    };
    warmSubs.add(update);
    if (warmBest) update(warmBest, null);
  }
  function render(root, {admin = false, edit = false, point = {}, nearby = null, me = null} = {}) {
    root._nearby = nearby;
    root.classList.add('pf');
    root.innerHTML = html({admin, edit});
    for (const el of root.querySelectorAll('[name]')) el.value = point[el.name] ?? '';
    fillNeeds(root, point.needs);
    drawMembers(root, membersOf(point).map(m => ({...m})));
    // Con mapa, el botón se llama «Ubicar en mapa»; sin mapa (formulario público) conserva «Enviar mi ubicación GPS».
    if (nearby) root.querySelector('.pf-gps').innerHTML = `${icons.pin}Ubicar en mapa`;
    root.querySelector('.pf-gps').onclick = () => locate(root);
    root.querySelector('.pf-gpsthumb').onclick = () => locate(root);
    root.querySelector('[name="gps"]').addEventListener('input', () => drawGpsThumb(root));
    setMe(root, me);
    drawGpsThumb(root);
    if (!admin) { warmGps(); gpsLive(root); }
    root.oninput = e => { const f = filters[e.target.dataset?.filter]; if (f) { const v = f(e.target.value); if (v !== e.target.value) e.target.value = v; } };
    root.onclick = e => {
      const b = e.target.closest('.pf-numbtn'); if (!b) return;
      const el = b.parentNode.querySelector('input');
      el.value = Math.min(999, Math.max(0, (parseInt(el.value, 10) || 0) + Number(b.dataset.step)));
      el.dispatchEvent(new Event('input', {bubbles: true}));
    };
    root.querySelector('.pf-members').addEventListener('input', () => {
      const box = root.querySelector('.pf-members'), rows = box.querySelectorAll('.pf-member'), last = rows[rows.length - 1];
      if (last && [...last.querySelectorAll('input')].some(el => el.value.trim())) box.insertAdjacentHTML('beforeend', memberRow({}, rows.length));
    });
    // «＋ Agregar otro contacto»: una fila nueva y el cursor en su nombre (también aparece sola al escribir en la última).
    root.querySelector('.pf-addmember').onclick = e => {
      e.preventDefault();
      const box = root.querySelector('.pf-members'), rows = box.querySelectorAll('.pf-member'), last = rows[rows.length - 1];
      if (!last || [...last.querySelectorAll('input')].some(el => el.value.trim())) box.insertAdjacentHTML('beforeend', memberRow({}, rows.length));
      const all = box.querySelectorAll('.pf-member'); all[all.length - 1].querySelector('[data-k="name"]').focus();
    };
    root.querySelector('.pf-members').onclick = e => {
      const b = e.target.closest('.pf-remove'); if (!b) return;
      const list = readMembers(root); list.splice([...root.querySelectorAll('.pf-remove')].indexOf(b), 1); drawMembers(root, list);
    };
  }
  // Datos para la API: campos de texto, «needs» armado desde las casillas y `members` sin filas vacías.
  function values(root) {
    const data = {};
    for (const el of root.querySelectorAll('[name]')) data[el.name] = el.type === 'checkbox' ? el.checked : el.value.trim();
    data.needs = readNeeds(root);
    data.members = readMembers(root).filter(m => Object.values(m).some(Boolean));
    return data;
  }
  const validPhone = s => (String(s ?? '').match(/\d/g) || []).length >= 7;
  const hasPhone = data => validPhone(data.coordinatorPhone) || (data.members || []).some(m => validPhone(m.phone));
  // Aviso de ubicación en vivo (página pública con código y panel de corredores): una casilla «Acepto compartir mi ubicación»
  // antes de compartirla si la persona no ha aceptado la versión vigente (el servidor guarda cuándo la aceptó; ver location.go). «Aceptar y activar»
  // solo se habilita con la casilla marcada. Devuelve true si aceptó. Si el texto cambia en algo importante, sube
  // locationNoticeVersion en location.go para que todos lo vuelvan a aceptar.
  function locationNotice() {
    let d = document.getElementById('pfLocNotice');
    if (!d) {
      if (!document.getElementById('pf-locnotice-style')) document.head.insertAdjacentHTML('beforeend', `<style id="pf-locnotice-style">dialog.pf-locnotice{width:min(440px,calc(calc(100vw / var(--text-zoom,1)) - 24px));max-height:calc(calc(100dvh / var(--text-zoom,1)) - 24px);padding:0;border:0;border-radius:16px;box-shadow:0 18px 50px #10263a66;color:#173542;font:15px system-ui,-apple-system,sans-serif}dialog.pf-locnotice::backdrop{background:#10263ab8}.pf-ln-box{padding:20px 20px 16px}.pf-ln-box h2{margin:0 0 10px;font-size:20px}.pf-ln-text{margin:0 0 14px;font-size:14px;line-height:1.5;color:#36505a}.pf-ln-ok{display:flex;gap:10px;align-items:flex-start;padding:12px;border:1.5px solid #b9d3d8;border-radius:10px;background:#f3faf9;font-weight:700;font-size:14px;line-height:1.35;cursor:pointer}.pf-ln-ok input{width:22px;height:22px;margin:0;flex:none;accent-color:#19745f}.pf-ln-actions{display:flex;gap:8px;justify-content:flex-end;margin-top:14px}.pf-ln-actions button{border:0;border-radius:10px;padding:11px 16px;font:inherit;font-weight:800;cursor:pointer;background:#19745f;color:#fff}.pf-ln-actions button:disabled{opacity:.45;cursor:not-allowed}.pf-ln-actions .pf-ln-no{background:#fff;color:#173542;border:1px solid #cad9d9}</style>`);
      document.body.insertAdjacentHTML('beforeend', `<dialog id="pfLocNotice" class="pf-locnotice" aria-labelledby="pfLocNoticeTitle"><form method="dialog" class="pf-ln-box"><h2 id="pfLocNoticeTitle">📍 Tu ubicación, por tu seguridad</h2><p class="pf-ln-text">Estás en una zona de desastre. Compartir tu ubicación en vivo puede salvarte: si algo pasa, se sabrá dónde estás y se podrá llegar a ti rápido. Puedes apagarla cuando quieras.</p><label class="pf-ln-ok"><input type="checkbox" name="ok" required> Por mi seguridad, acepto que mi ubicación se use y se comparta mientras esté activa.</label><div class="pf-ln-actions"><button type="submit" class="pf-ln-no" value="no" formnovalidate>Ahora no</button><button type="submit" value="yes" disabled>Aceptar y activar</button></div></form></dialog>`);
      d = document.getElementById('pfLocNotice');
      const ok = d.querySelector('[name=ok]');
      ok.onchange = () => { d.querySelector('[value=yes]').disabled = !ok.checked; };
    }
    const ok = d.querySelector('[name=ok]');
    ok.checked = false;
    d.querySelector('[value=yes]').disabled = true;
    d.returnValue = '';
    d.showModal();
    return new Promise(resolve => d.addEventListener('close', () => resolve(d.returnValue === 'yes' && ok.checked), {once: true}));
  }

  // Cola sin conexión (página pública y Runners). Al guardar un punto sin internet (o si el servidor no responde) queda
  // en este navegador (localStorage), sus datos se copian al portapapeles y se envía solo al volver la conexión, al abrir la
  // página o cada minuto. Cada punto lleva clientRef: si el servidor ya lo recibió, el reenvío no crea otro.
  // Si el servidor lo rechaza (dato inválido), queda en la cola con el motivo para corregirlo o quitarlo.
  const QKEY = 'pointQueue';
  let qSend = null, qSent = null, qBusy = false;
  const qLoad = () => { try { const l = JSON.parse(localStorage.getItem(QKEY) || '[]'); return Array.isArray(l) ? l : []; } catch { return []; } };
  const qSave = l => { try { localStorage.setItem(QKEY, JSON.stringify(l)); } catch {} };
  const newRef = () => 'q-' + (crypto.randomUUID ? crypto.randomUUID() : Date.now().toString(36) + Math.random().toString(36).slice(2));
  function queuePoint(record) {
    const l = qLoad();
    if (!l.some(x => x.ref === record.clientRef)) l.push({ref: record.clientRef, record, at: new Date().toISOString(), error: ''});
    qSave(l);
    try { navigator.clipboard?.writeText(JSON.stringify(record, null, 2)).catch(() => {}); } catch {}
    drawQueue();
  }
  // Envía el punto; sin conexión (o 502–504) lo deja en la cola. Devuelve {queued:true} o {res}.
  async function sendOrQueue(record, send) {
    record.clientRef ||= newRef();
    if (navigator.onLine === false) { queuePoint(record); enterOffline(true); return {queued: true}; }
    let r;
    try { r = await send(record); } catch { queuePoint(record); enterOffline(true); return {queued: true}; }
    if (r.status >= 502 && r.status <= 504) { queuePoint(record); enterOffline(true); return {queued: true}; }
    return {res: r};
  }
  async function flushQueue() {
    if (qBusy || !qSend) return;
    if (!qLoad().length) { drawQueue(); return; }
    qBusy = true; drawQueue();
    let sent = 0;
    for (const item of qLoad()) {
      if (item.error) continue;
      let r;
      try { r = await qSend(item.record); } catch { break; }
      if (r.ok) { qSave(qLoad().filter(x => x.ref !== item.ref)); sent++; continue; }
      if (r.status >= 500 || r.status === 401 || r.status === 429) break;
      const body = await r.json().catch(() => ({})), l = qLoad(), x = l.find(x => x.ref === item.ref);
      if (x) { x.error = body.error || 'Error HTTP ' + r.status; qSave(l); }
    }
    qBusy = false; drawQueue();
    if (sent && qSent) qSent(sent);
  }
  function queueStyle() {
    if (document.getElementById('pf-queue-style')) return;
    document.head.insertAdjacentHTML('beforeend', `<style id="pf-queue-style">.pfq-chip{position:fixed;left:50%;bottom:calc(14px + env(safe-area-inset-bottom));transform:translateX(-50%);z-index:1500;display:inline-flex;align-items:center;gap:8px;border:0;border-radius:999px;background:#10263a;color:#fff;padding:10px 16px;font:800 14px Inter,system-ui,sans-serif;box-shadow:0 8px 24px #10263a66;cursor:pointer}.pfq-chip[hidden]{display:none}.pfq-chip .err{background:#c62828;border-radius:999px;padding:1px 8px;font-size:12px}.pfq-dlg{border:0;border-radius:16px;padding:0;width:min(460px,calc(100% - 24px));max-height:calc(88dvh / var(--text-zoom,1));color:#173c46;font:15px Inter,system-ui,sans-serif;box-shadow:0 20px 60px #10263a66}.pfq-dlg::backdrop{background:#10263ab8}.pfq-box{padding:18px;display:flex;flex-direction:column;gap:10px}.pfq-box h2{margin:0;font-size:18px}.pfq-box p{margin:0;color:#667982;font-size:13px;line-height:1.45}.pfq-list{list-style:none;margin:0;padding:0;display:flex;flex-direction:column;gap:6px;max-height:calc(50dvh / var(--text-zoom,1));overflow:auto}.pfq-list li{border:1px solid #e1e9ea;border-radius:10px;padding:9px 11px;display:flex;flex-direction:column;gap:4px}.pfq-list b{font-size:14px}.pfq-list small{color:#667982}.pfq-list .e{color:#a33527;font-weight:700}.pfq-row{display:flex;gap:6px;flex-wrap:wrap}.pfq-row button{border:1px solid #cad9d9;background:#fff;color:#173c46;border-radius:8px;padding:5px 10px;font:700 12px Inter,system-ui,sans-serif;cursor:pointer}.pfq-row .del{color:#a33527;border-color:#e3b9b0}.pfq-acts{display:flex;gap:8px}.pfq-acts button{flex:1;border:0;border-radius:10px;padding:12px;font:800 15px Inter,system-ui,sans-serif;cursor:pointer;background:#eef3f3;color:#173c46}.pfq-acts .go{background:#19745f;color:#fff}</style>`);
  }
  function drawQueue() {
    if (!qSend) return;
    queueStyle();
    let chip = document.getElementById('pfQueue');
    if (!chip) { chip = document.createElement('button'); chip.type = 'button'; chip.id = 'pfQueue'; chip.className = 'pfq-chip'; document.body.appendChild(chip); chip.onclick = openQueue; }
    const l = qLoad(), bad = l.filter(x => x.error).length;
    chip.hidden = !l.length;
    chip.innerHTML = `📤 ${qBusy ? 'Enviando…' : `${l.length} ${l.length === 1 ? 'punto por enviar' : 'puntos por enviar'}`}${bad ? ` <span class="err">${bad} con error</span>` : ''}`;
    if (document.getElementById('pfQueueDlg')?.open) fillQueue();
  }
  function fillQueue() {
    const d = document.getElementById('pfQueueDlg'), l = qLoad();
    d.querySelector('.pfq-list').innerHTML = l.length ? l.map((x, i) => `<li><b>${esc(x.record.area || 'Sin zona')}${x.record.place ? ' · ' + esc(x.record.place) : ''}</b><small>Guardado ${esc(new Date(x.at).toLocaleTimeString('es-PA', {hour: 'numeric', minute: '2-digit'}))}${x.record.coordinatorPhone ? ' · ' + esc(x.record.coordinatorPhone) : ''}</small>${x.error ? `<span class="e">No se pudo enviar: ${esc(x.error)}</span>` : ''}<div class="pfq-row"><button type="button" data-qcopy="${i}">Copiar datos</button>${x.error ? `<button type="button" data-qretry="${i}">Reintentar</button>` : ''}<button type="button" class="del" data-qdel="${i}">Quitar</button></div></li>`).join('') : '<li><small>No hay puntos por enviar.</small></li>';
  }
  function openQueue() {
    queueStyle();
    let d = document.getElementById('pfQueueDlg');
    if (!d) {
      document.body.insertAdjacentHTML('beforeend', '<dialog id="pfQueueDlg" class="pfq-dlg" aria-labelledby="pfQueueTitle"><div class="pfq-box"><h2 id="pfQueueTitle">Puntos por enviar</h2><p>Se guardaron en este teléfono porque no había conexión. Se envían solos cuando vuelve internet.</p><ul class="pfq-list"></ul><div class="pfq-acts"><button type="button" data-qclose>Cerrar</button><button type="button" class="go" data-qsend>Enviar ahora</button></div></div></dialog>');
      d = document.getElementById('pfQueueDlg');
      d.addEventListener('click', async e => {
        const b = e.target.closest('button');
        if (!b) return;
        const l = qLoad();
        if (b.dataset.qclose !== undefined) d.close();
        else if (b.dataset.qsend !== undefined) { l.forEach(x => { x.error = ''; }); qSave(l); await flushQueue(); fillQueue(); }
        else if (b.dataset.qcopy) { try { await navigator.clipboard.writeText(JSON.stringify(l[+b.dataset.qcopy].record, null, 2)); b.textContent = 'Copiado'; } catch { b.textContent = 'No se pudo copiar'; } }
        else if (b.dataset.qretry) { l[+b.dataset.qretry].error = ''; qSave(l); await flushQueue(); fillQueue(); }
        else if (b.dataset.qdel && confirm('¿Quitar este punto de la cola? No se enviará.')) { l.splice(+b.dataset.qdel, 1); qSave(l); drawQueue(); fillQueue(); }
      });
    }
    fillQueue();
    d.showModal();
  }

  // Modo sin conexión: al perder la conexión (o si un envío falla por falta de red) una pantalla cubre toda la página.
  // Ahí solo se ven los puntos guardados en el teléfono: tocar uno lo abre para editarlo, «+ Agregar otro punto» y
  // «Reintentar envío». No se sale hasta que hay conexión y todos se enviaron bien. Si había un formulario abierto,
  // lo escrito pasa a la pantalla sin conexión para no perderlo.
  let offOn = false, offEdit = null;
  function offlineStyle() {
    if (document.getElementById('pf-offline-style')) return;
    document.head.insertAdjacentHTML('beforeend', `<style id="pf-offline-style">.pfo{position:fixed;inset:0;z-index:5000;display:flex;flex-direction:column;background:#f2f6f5;color:#173c46;font:15px Inter,system-ui,sans-serif}.pfo[hidden]{display:none}.pfo-head{flex:none;display:flex;align-items:center;gap:10px;padding:max(12px,env(safe-area-inset-top)) 16px 12px;background:#10263a;color:#fff}.pfo-head h2{margin:0;font-size:18px;flex:1}.pfo-dot{width:10px;height:10px;border-radius:50%;background:#f0a54a;box-shadow:0 0 0 4px #f0a54a33}.pfo-badge{min-width:26px;text-align:center;border-radius:999px;background:#ed7654;color:#fff;font-weight:800;padding:3px 9px;font-size:14px}.pfo-body{flex:1;overflow:auto;padding:14px 16px 18px}.pfo-body>p{margin:0 0 12px;color:#4b6670;font-size:14px;line-height:1.45}.pfo-list{list-style:none;margin:0 0 12px;padding:0;display:flex;flex-direction:column;gap:8px}.pfo-list button{width:100%;text-align:left;border:1px solid #d6e1e0;border-radius:12px;background:#fff;padding:12px 14px;font:inherit;color:#173c46;cursor:pointer;display:flex;flex-direction:column;gap:3px;box-shadow:0 1px 3px #17354214}.pfo-list b{font-size:15px}.pfo-list small{color:#667982}.pfo-list .e{color:#a33527;font-weight:700;font-size:13px}.pfo-add{width:100%;border:2px dashed #9fb6bb;border-radius:12px;background:#fff;color:#19745f;font:800 15px Inter,system-ui,sans-serif;padding:13px;cursor:pointer}.pfo-foot{flex:none;display:flex;flex-direction:column;gap:6px;padding:10px 16px max(12px,env(safe-area-inset-bottom));background:#fff;border-top:1px solid #dce5e4}.pfo-foot .row{display:flex;gap:8px}.pfo-foot button{flex:1;border:0;border-radius:12px;padding:14px;font:800 16px Inter,system-ui,sans-serif;cursor:pointer}.pfo-go{background:#19745f;color:#fff}.pfo-plain{background:#eef3f3;color:#173c46}.pfo-status{margin:0;text-align:center;font-size:13px;color:#4b6670;min-height:18px}.pfo-del{margin-top:14px;width:100%;border:1px solid #e3b9b0;border-radius:10px;background:#fff;color:#a33527;font:700 14px Inter,system-ui,sans-serif;padding:11px;cursor:pointer}.pfo-toast{position:fixed;left:50%;bottom:calc(20px + env(safe-area-inset-bottom));transform:translateX(-50%);z-index:5100;background:#19745f;color:#fff;border-radius:999px;padding:10px 18px;font:800 14px Inter,system-ui,sans-serif;box-shadow:0 8px 24px #0004}</style>`);
  }
  function offlineEl() {
    let o = document.getElementById('pfOffline');
    if (o) return o;
    offlineStyle();
    document.body.insertAdjacentHTML('beforeend', '<div id="pfOffline" class="pfo" role="dialog" aria-modal="true" aria-labelledby="pfoTitle" hidden><div class="pfo-head"><span class="pfo-dot" aria-hidden="true"></span><h2 id="pfoTitle">Sin conexión</h2><span class="pfo-badge" id="pfoBadge" title="Puntos por enviar">0</span></div><div class="pfo-body" id="pfoBody"></div><div class="pfo-foot" id="pfoFoot"></div></div>');
    o = document.getElementById('pfOffline');
    o.addEventListener('click', offlineClick);
    return o;
  }
  function lockPage(on) { for (const el of document.body.children) if (el.id !== 'pfOffline' && !el.classList.contains('pfo-toast')) el.inert = on; }
  // saved: el punto del formulario ya quedó en la cola (al guardar sin conexión), así que no se copia como borrador.
  const addFormOpen = () => [...document.querySelectorAll('.pf')].some(r => r.id !== 'pfoForm' && r.closest('.modalbg.open,dialog[open]'));
  function enterOffline(saved = false) {
    if (!qSend) return;
    const o = offlineEl();
    if (!offOn) {
      // Lo que se estaba escribiendo en un formulario abierto pasa aquí como punto nuevo.
      const live = [...document.querySelectorAll('.pf')].find(r => r.id !== 'pfoForm' && r.closest('.modalbg.open,dialog[open]'));
      if (live) {
        const draft = values(live);
        if (!saved && Object.entries(draft).some(([k, v]) => k === 'members' ? v.length : String(v ?? '').trim())) offEdit = {draft};
        const host = live.closest('.modalbg.open,dialog[open]');
        if (host.tagName === 'DIALOG') host.close(); else host.classList.remove('open');
      }
    }
    offOn = true;
    o.hidden = false;
    lockPage(true);
    drawOffline();
  }
  function leaveOffline(msg) {
    offOn = false; offEdit = null;
    offlineEl().hidden = true;
    lockPage(false);
    if (msg) { const t = document.createElement('div'); t.className = 'pfo-toast'; t.textContent = msg; document.body.appendChild(t); setTimeout(() => t.remove(), 3500); }
  }
  function drawOffline(status = '') {
    const l = qLoad(), body = document.getElementById('pfoBody'), foot = document.getElementById('pfoFoot');
    document.getElementById('pfoBadge').textContent = l.length;
    if (offEdit) {
      const existing = offEdit.ref && l.find(x => x.ref === offEdit.ref);
      body.innerHTML = `<p>${existing ? 'Corrige el punto y guárdalo de nuevo en el teléfono.' : 'Nuevo punto: se guarda en el teléfono hasta que haya conexión.'}</p><div id="pfoForm"></div>${existing ? '<button type="button" class="pfo-del" data-o="del">Quitar este punto</button>' : ''}<p class="pfo-status" id="pfoFormMsg" role="status"></p>`;
      render(document.getElementById('pfoForm'), {point: existing ? existing.record : offEdit.draft || {}});
      foot.innerHTML = '<div class="row"><button type="button" class="pfo-plain" data-o="back">Volver</button><button type="button" class="pfo-go" data-o="save">Guardar en el teléfono</button></div>';
      body.scrollTop = 0;
      return;
    }
    body.innerHTML = `<p>No hay conexión. Puedes seguir agregando puntos: quedan guardados en este teléfono y se envían cuando vuelva internet. Toca uno para editarlo.</p><ul class="pfo-list">${l.map(x => `<li><button type="button" data-o="edit" data-ref="${esc(x.ref)}"><b>${esc(x.record.area || 'Sin zona')}${x.record.place ? ' · ' + esc(x.record.place) : ''}</b><small>${esc(new Date(x.at).toLocaleTimeString('es-PA', {hour: 'numeric', minute: '2-digit'}))}${x.record.coordinatorPhone ? ' · ' + esc(x.record.coordinatorPhone) : ''}${x.record.needs ? ' · ' + esc(x.record.needs) : ''}</small>${x.error ? `<span class="e">El servidor no lo aceptó: ${esc(x.error)}. Tócalo para corregirlo.</span>` : ''}</button></li>`).join('')}</ul><button type="button" class="pfo-add" data-o="new">+ Agregar otro punto</button>`;
    foot.innerHTML = `<p class="pfo-status" id="pfoStatus" role="status">${esc(status)}</p><button type="button" class="pfo-go" data-o="check">Reintentar envío</button>`;
  }
  // Reintentar envío: si hay conexión se envía todo; solo cuando la cola queda vacía se sale.
  let checking = false;
  async function checkOnline() {
    if (checking) return;
    checking = true;
    drawOffline('Intentando enviar…');
    const ok = await fetch('/health', {cache: 'no-store'}).then(r => r.ok).catch(() => false);
    if (!ok) { checking = false; drawOffline('Todavía sin conexión. Intenta de nuevo en un momento.'); return; }
    drawOffline('Conectado. Enviando…');
    const before = qLoad().length;
    await flushQueue();
    checking = false;
    const l = qLoad();
    if (!l.length) { leaveOffline(before ? `Conectado: ${before === 1 ? 'se envió 1 punto' : `se enviaron ${before} puntos`}.` : 'Conectado.'); return; }
    drawOffline(l.some(x => x.error) ? 'Hay puntos que el servidor no aceptó: tócalos para corregirlos.' : 'El servidor no respondió. Intenta de nuevo en un momento.');
  }
  function offlineClick(e) {
    const b = e.target.closest('[data-o]');
    if (!b) return;
    const act = b.dataset.o;
    if (act === 'check') checkOnline();
    else if (act === 'new') { offEdit = {draft: {}}; drawOffline(); }
    else if (act === 'edit') { offEdit = {ref: b.dataset.ref}; drawOffline(); }
    else if (act === 'back') { offEdit = null; drawOffline(); }
    else if (act === 'del') {
      if (!confirm('¿Quitar este punto? No se enviará.')) return;
      qSave(qLoad().filter(x => x.ref !== offEdit.ref)); offEdit = null; drawQueue(); drawOffline();
    } else if (act === 'save') {
      const record = values(document.getElementById('pfoForm'));
      if (!hasPhone(record)) { document.getElementById('pfoFormMsg').textContent = 'Agrega al menos un contacto con teléfono (7 dígitos o más).'; return; }
      const l = qLoad(), x = offEdit.ref && l.find(x => x.ref === offEdit.ref);
      if (x) { record.clientRef = x.ref; x.record = record; x.error = ''; qSave(l); }
      else { record.clientRef = newRef(); l.push({ref: record.clientRef, record, at: new Date().toISOString(), error: ''}); qSave(l); }
      offEdit = null; drawQueue(); drawOffline('Guardado en el teléfono.');
    }
  }
  // La página dice cómo enviar (send) y qué hacer tras enviar (sent, p. ej. recargar la lista).
  function setQueue({send, sent} = {}) {
    qSend = send; qSent = sent || null;
    drawQueue();
    if (navigator.onLine !== false) flushQueue();
  }
  if (typeof addEventListener === 'function' && typeof document !== 'undefined' && document.addEventListener) {
    addEventListener('online', () => { if (offOn) checkOnline(); else flushQueue(); });
    // Solo si estás agregando un punto (formulario abierto); navegando sin conexión no se bloquea nada.
    addEventListener('offline', () => { if (addFormOpen()) enterOffline(); });
    document.addEventListener('visibilitychange', () => { if (document.visibilityState === 'visible') flushQueue(); });
    setInterval(() => { if (qSend && qLoad().some(x => !x.error)) flushQueue(); }, 60000);
  }
  // Solicitudes de cambio (changes.go): lo que un agente de otro equipo (o sin equipo) propone cambiar en una necesidad
  // que tú actualizaste por última vez; tú estuviste ahí, así que aceptas o rechazas. showSent: también las que enviaste
  // (Command). load() devuelve {forMe, sent}; decide(id, 'accept'|'reject').
  function changeRequests(box, {load, decide, after, showSent} = {}) {
    if (!box) return {refresh() {}};
    if (!document.getElementById('crqStyle')) {
      const st = document.createElement('style');
      st.id = 'crqStyle';
      st.textContent = '.crqbox{margin:10px 0;display:grid;gap:8px}.crqbox[hidden]{display:none}.crqbox h3{margin:0;font-size:15px;color:#173c46}.crqbox h3 b{background:#ed7654;color:#fff;border-radius:10px;padding:1px 7px;font-size:12px}.crq-help{margin:0;font-size:12px;color:#4b6670;line-height:1.4}.crq{border:1.5px solid #f0b9a8;background:#fff8f5;border-radius:11px;padding:10px 12px;font-size:13px;color:#173542}.crq.sent{border-color:#d5e0e2;background:#f7faf9}.crq.sent.accepted{border-color:#9fd3b5}.crq.sent.rejected{border-color:#e7b0b0}.crq-who{font-size:12px;color:#4b6670}.crq-need{font-weight:750;margin:3px 0}.crq ul{margin:4px 0 0;padding-left:18px}.crq li{margin:2px 0;overflow-wrap:anywhere}.crq s{color:#8a9aa0}.crq-acts{display:flex;gap:8px;margin-top:8px}.crq-acts button{flex:1;border:0;border-radius:9px;padding:10px;font:inherit;font-weight:800;cursor:pointer}.crq-yes{background:#19745f;color:#fff}.crq-no{background:#fff;color:#a83c1f;border:1.5px solid #e7b0b0!important}.crq-sent summary{cursor:pointer;font-weight:750;font-size:13px;color:#126e7c;padding:4px 0}.crq-sent .crq{margin-top:6px}';
      document.head.appendChild(st);
    }
    const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
    let data = {forMe: [], sent: []}, sentOpen = false;
    const line = c => `<li><b>${esc(c.label)}:</b> <s>${esc(c.before || '—')}</s> → <b>${esc(c.after || '—')}</b></li>`;
    const where = r => esc([r.campCode, r.campArea].filter(Boolean).join(' · '));
    function draw() {
      const forMe = data.forMe || [], sent = showSent ? data.sent || [] : [];
      box.hidden = !forMe.length && !sent.length;
      const waiting = sent.filter(r => r.status === 'pending').length;
      box.innerHTML = (forMe.length ? `<h3>📝 Cambios propuestos <b>${forMe.length}</b></h3><p class="crq-help">Alguien de otro equipo propone cambiar algo que tú actualizaste. Tú estuviste ahí: acepta solo si es correcto.</p>` +
        forMe.map(r => `<div class="crq" data-crq="${esc(r.id)}"><div class="crq-who"><b>${esc(r.fromName)}</b> · ${r.fromTeam ? 'Equipo ' + esc(r.fromTeam) : 'sin equipo'} · ${where(r)}</div><div class="crq-need">${esc(r.need)}</div><ul>${(r.changes || []).map(line).join('')}</ul><div class="crq-acts"><button type="button" class="crq-yes" data-crq-yes>✓ Aceptar</button><button type="button" class="crq-no" data-crq-no>✗ Rechazar</button></div></div>`).join('') : '') +
        (sent.length ? `<details class="crq-sent"${sentOpen ? ' open' : ''}><summary>📤 Cambios que pediste${waiting ? ` · ${waiting} esperando respuesta` : ''}</summary>` +
        sent.map(r => `<div class="crq sent ${esc(r.status)}"><div class="crq-who">${({pending: '⏳ Esperando a', accepted: '✓ Lo aceptó', rejected: '✗ Lo rechazó'})[r.status] || esc(r.status)} <b>${esc(r.toName)}</b> · ${where(r)}</div><div class="crq-need">${esc(r.need)}</div><ul>${(r.changes || []).map(line).join('')}</ul></div>`).join('') + '</details>' : '');
    }
    box.addEventListener('toggle', e => { if (e.target.matches?.('.crq-sent')) sentOpen = e.target.open; }, true);
    box.addEventListener('click', async e => {
      const yes = e.target.closest('[data-crq-yes]'), no = e.target.closest('[data-crq-no]');
      if (!yes && !no) return;
      const card = e.target.closest('[data-crq]'), btns = card.querySelectorAll('button');
      btns.forEach(b => b.disabled = true);
      try { await decide(card.dataset.crq, yes ? 'accept' : 'reject'); await refresh(); if (after) after(!!yes); }
      catch (err) { alert('No se pudo: ' + err.message); btns.forEach(b => b.disabled = false); }
    });
    async function refresh() { try { data = await load(); draw(); } catch {} }
    refresh();
    setInterval(() => { if (!document.hidden) refresh(); }, 30000);
    return {refresh};
  }
  // Historial de cambios (audit.go): cada cambio de un punto y de sus necesidades (runs, empaque, fotos, despacho,
  // acopio) con quién lo hizo, su equipo, desde qué página y cuándo, del más reciente al más antiguo. Un <details
  // data-trail="ID del punto"> con un <div class="trail"> se llena al abrirlo; urlFor(id) da la dirección según el panel.
  function trailHtml(list) {
    const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
    const cut = v => { v = String(v ?? ''); return v.length > 240 ? v.slice(0, 240) + '…' : v; };
    const when = t => { const d = new Date(t); return isNaN(d) ? '' : d.toLocaleString('es-PA', {day: 'numeric', month: 'short', hour: 'numeric', minute: '2-digit'}); };
    if (!list || !list.length) return '<p class="tr-none">Sin cambios registrados todavía.</p>';
    return '<ol class="trlist">' + list.map(e => `<li class="tr-${e.kind === 'punto' ? 'point' : e.kind === 'acopio' ? 'center' : 'need'}"><div class="tr-head"><b>${esc(e.byName)}</b>${e.byTeam ? ` · Equipo ${esc(e.byTeam)}` : e.byId && e.byId !== 'admin' ? ' · sin equipo' : ''}${e.via ? ` · ${esc(e.via)}` : ''} · <time datetime="${esc(e.at)}">${esc(when(e.at))}</time></div><div class="tr-what">${e.kind === 'punto' ? '📍 Punto' : e.kind === 'acopio' ? '🏬 Acopio' : '📦 ' + esc(cut(e.label))} · ${e.action === 'creó' ? 'creado' : 'cambio'}</div><ul>${(e.changes || []).map(c => `<li><b>${esc(c.label)}:</b> ${c.before ? `<s>${esc(cut(c.before))}</s> → ` : ''}${c.after ? esc(cut(c.after)) : '<i>(vacío)</i>'}</li>`).join('')}</ul></li>`).join('') + '</ol>';
  }
  function watchTrails(urlFor) {
    if (!document.getElementById('trailStyle')) {
      const st = document.createElement('style');
      st.id = 'trailStyle';
      st.textContent = '.trlist{list-style:none;margin:8px 0 0;padding:0;display:grid;gap:8px;font-size:13px;color:#173542}.trlist>li{border-left:3px solid #126e7c;background:#f7faf9;border-radius:0 9px 9px 0;padding:7px 10px}.trlist>li.tr-point{border-left-color:#ed7654}.trlist>li.tr-center{border-left-color:#7b3fbf}.tr-head{font-size:12px;color:#4b6670}.tr-head b{color:#173542}.tr-what{font-weight:750;margin:2px 0}.trlist ul{margin:2px 0 0;padding-left:18px}.trlist ul li{overflow-wrap:anywhere;margin:1px 0}.trlist s{color:#8a9aa0}.tr-none{color:#4b6670;font-size:13px}';
      document.head.appendChild(st);
    }
    document.addEventListener('toggle', async e => {
      const d = e.target;
      if (!d.matches?.('details[data-trail]') || !d.open || !d.dataset.trail) return;
      const box = d.querySelector('.trail');
      if (!box) return;
      box.textContent = 'Cargando…';
      try {
        const r = await fetch(urlFor(d.dataset.trail), {credentials: 'same-origin', cache: 'no-store'});
        if (!r.ok) throw Error();
        box.innerHTML = trailHtml(await r.json());
      } catch { box.textContent = 'No se pudo cargar el historial.'; }
    }, true);
  }
  window.PointForm = {trailHtml, watchTrails, changeRequests, needChipsHtml, verifiedBadge, verifiedTitle, BRANDED, DATA_NEED, brandOf, brandBadge, dataNet, netBadge, needBase, needChip, render, values, html, nearest, meters, lastAdded, linkCoords, needsHtml, fillNeeds, readNeeds, membersOf, parseNeeds, needsText, filters, validPhone, hasPhone, NEEDS, icons, PACKING, packIcon, packingTotals, packingHtml, openPacking, prepareMedia, uploadMedia, mediaThumbs, setMe, locationNotice, sendOrQueue, setQueue, queuePoint, flushQueue, enterOffline, SERVICES, serviceIcon, serviceStatus, serviceScheduled};
})();
