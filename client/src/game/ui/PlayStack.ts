// PlayStack represents the stack in the middle of the table for each suit
// It is composed of LayoutChild objects

import Konva from 'konva';
import * as variantRules from '../rules/variant';
import { STACK_BASE_RANK } from '../types/constants';
import globals from './globals';
import HanabiCard from './HanabiCard';
import { animate } from './konvaHelpers';
import LayoutChild from './LayoutChild';

export default class PlayStack extends Konva.Group {
  addChild(layoutChild: LayoutChild) {
    const pos = layoutChild.getAbsolutePosition();
    this.add(layoutChild as any);
    layoutChild.setAbsolutePosition(pos);
    this.doLayout();
  }

  doLayout() {
    const lh = this.height();

    for (const layoutChild of this.children.toArray() as LayoutChild[]) {
      const scale = lh / layoutChild.height();
      const card = layoutChild.children[0] as unknown as HanabiCard;
      const stackBase = card.state.rank === STACK_BASE_RANK;
      const opacity = (
        // Hide cards in "Throw It in a Hole" variants
        variantRules.isThrowItInAHole(globals.variant)
        && !globals.replay // Revert to the normal behavior for replays
        && !stackBase // We want the stack bases to always be visible
      ) ? 0 : 1;

      // Animate the card leaving the hand to the play stacks
      // (tweening from the hand to the discard pile is handled in
      // the "CardLayout" object)
      card.startedTweening();
      animate(layoutChild, {
        duration: 0.8,
        x: 0,
        y: 0,
        scale,
        rotation: 0,
        opacity,
        easing: Konva.Easings.EaseOut,
        onFinish: () => {
          if (!layoutChild || !card || !card.parent) {
            return;
          }
          if (layoutChild.tween !== null) {
            layoutChild.tween.destroy();
            layoutChild.tween = null;
          }
          card.finishedTweening();
          layoutChild.checkSetDraggable();
          this.hideCardsUnderneathTheTopCard();
        },
      });
    }
  }

  hideCardsUnderneathTheTopCard() {
    const stackLength = this.children.length;

    for (let i = 0; i < stackLength; i++) {
      const layoutChild = this.children[i] as unknown as LayoutChild;
      if (layoutChild.tween !== null) {
        // Don't hide anything if one of the cards on the stack is still tweening
        return;
      }
    }

    // Hide all of the cards
    for (let i = 0; i < stackLength - 1; i++) {
      this.children[i].hide();
    }

    // Show the top card
    if (stackLength > 0) {
      this.children[stackLength - 1].show();
    }
  }

  getLastPlayedRank() {
    // The PlayStack will always have at least 1 element in it (the "stack base" card)
    const topLayoutChild = this.children[this.children.length - 1];
    const topCard = topLayoutChild.children[0] as HanabiCard;
    return topCard.state.rank;
  }
}
