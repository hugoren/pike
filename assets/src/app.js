import { h, app } from 'hyperapp';
import 'whatwg-fetch';
import moment from 'moment';

import './global.sss';
import './app.sss';
import Performance from './performance';
import Directors from './directors';
import AdminToken from './admin-token';
import BlockIP from './block-ip';
import Cached from './cached';
import {
  state,
  actions,
  getStats,
} from './actions';

const views = {
  default: 'default',
  adminToken: 'adminToken',
  performance: 'performance',
  blockIP: 'blockIP',
  cached: 'cached',
};

let toggleCount = 0;

let refreshStatsInterval = null;
const refreshStats = () => {
  getStats().then((data) => {
    main.setPerformance(data);
  });
}

const init = () => {
  getStats().then((data) => {
    main.setLaunchedAt(data.startedAt);
    main.setVersion(data.version);
    refreshStats();
    refreshStatsInterval = setInterval(refreshStats, 60 * 1000);
  }).catch((res) => {
    if (res.status === 401) {
      main.changeView(views.adminToken);
    }
  });
};

const view = (state, actions) => {
  const currentView = state.view;
  const getNav = (view, name) => {
    return <li><a
      href="javascript:;"
      class={currentView == view ? "active": ""}
      onclick={() => {
        toggleCount++;
        actions.changeView(view);
      }}
    >
      {name} 
    </a></li> 
  }
  return <div>
    <nav class="navBar">Pike
      <span class="version">({state.version})</span>
      <ul>
        {getNav(views.default, "Directors")}
        {getNav(views.performance, "Performance")}
        {getNav(views.blockIP, "Block IP List")}
        {getNav(views.cached, "Cached List")}
      </ul>
      {
        state.uptime &&
        <div
          class="launthedAt grayColor"
          title={state.launchedAt}
        >
          launthed at:
          <span class="mleft5">{state.uptime}</span>
        </div>
      }
    </nav>
    {
      currentView === views.adminToken && <AdminToken state={state} />
    }
    {
      currentView === views.performance && <Performance state={state} />
    }
    {
      currentView === views.default && <Directors state={state} actions={actions} toggleCount={toggleCount} />
    }
    {
      currentView === views.blockIP && <BlockIP state={state} actions={actions} toggleCount={toggleCount} />
    }
    {
      currentView === views.cached && <Cached state={state} actions={actions} toggleCount={toggleCount} />
    }
  </div>
};

const main = app(state, actions, view, document.body)

init();
