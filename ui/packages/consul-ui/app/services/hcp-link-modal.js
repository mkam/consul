/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: BUSL-1.1
 */

import Service from '@ember/service';
import { tracked } from '@glimmer/tracking';

export default class HcpLinkModalService extends Service {
  @tracked isModalVisible = false;
  @tracked resourceId = null;

  show(hcpLinkData) {
    this.isModalVisible = true;
  }

  hide() {
    this.isModalVisible = false;
  }
  setResourceId(resourceId) {
    this.resourceId = resourceId;
  }
}
