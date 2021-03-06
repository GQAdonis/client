import * as React from 'react'
import Banner from '../bottom-banner/container'
import HeaderArea from '../header-area/container.native'
import InputArea from '../input-area/container'
import ListArea from '../list-area/container'
import * as Kb from '../../../common-adapters'
import {LayoutEvent} from '../../../common-adapters/box'
import * as Styles from '../../../styles'
import {Props} from '.'
import ThreadLoadStatus from '../load-status/container'
import PinnedMessage from '../pinned-message/container'
import {GatewayDest} from 'react-gateway'
import InvitationToBlock from '../../blocking/invitation-to-block'

const Offline = () => (
  <Kb.Box
    style={{
      ...Styles.globalStyles.flexBoxCenter,
      backgroundColor: Styles.globalColors.greyDark,
      paddingBottom: Styles.globalMargins.tiny,
      paddingLeft: Styles.globalMargins.medium,
      paddingRight: Styles.globalMargins.medium,
      paddingTop: Styles.globalMargins.tiny,
      width: '100%',
    }}
  >
    <Kb.Text center={true} type="BodySmallSemibold">
      Couldn't load all chat messages due to network connectivity. Retrying...
    </Kb.Text>
  </Kb.Box>
)

const Conversation = React.memo((props: Props) => {
  const [maxInputArea, setMaxInputArea] = React.useState<number | undefined>(undefined)
  const onLayout = React.useCallback((e: LayoutEvent) => {
    setMaxInputArea(e.nativeEvent.layout.height)
  }, [])
  const innerComponent = (
    <Kb.BoxGrow onLayout={onLayout}>
      <Kb.Box2 direction="vertical" fullWidth={true} style={styles.innerContainer}>
        <ThreadLoadStatus conversationIDKey={props.conversationIDKey} />
        <PinnedMessage conversationIDKey={props.conversationIDKey} />
        <ListArea
          scrollListDownCounter={props.scrollListDownCounter}
          scrollListToBottomCounter={props.scrollListToBottomCounter}
          scrollListUpCounter={props.scrollListUpCounter}
          onFocusInput={props.onFocusInput}
          conversationIDKey={props.conversationIDKey}
        />
        {props.showLoader && <Kb.LoadingLine />}
      </Kb.Box2>
      <InvitationToBlock conversationID={props.conversationIDKey} />
      <Banner conversationIDKey={props.conversationIDKey} />
      <InputArea
        focusInputCounter={props.focusInputCounter}
        jumpToRecent={props.jumpToRecent}
        onRequestScrollDown={props.onRequestScrollDown}
        onRequestScrollToBottom={props.onRequestScrollToBottom}
        onRequestScrollUp={props.onRequestScrollUp}
        conversationIDKey={props.conversationIDKey}
        maxInputArea={maxInputArea}
      />
    </Kb.BoxGrow>
  )
  return (
    <Kb.Box style={styles.innerContainer}>
      <Kb.Box2 direction="vertical" fullWidth={true} fullHeight={true}>
        {props.threadLoadedOffline && <Offline />}
        {Styles.isTablet ? (
          innerComponent
        ) : (
          <>
            <HeaderArea conversationIDKey={props.conversationIDKey} />
            {innerComponent}
          </>
        )}
      </Kb.Box2>
      <GatewayDest name="convOverlay" component={Kb.Box} />
    </Kb.Box>
  )
})

const styles = Styles.styleSheetCreate(
  () =>
    ({
      innerContainer: {
        flex: 1,
        position: 'relative',
      },
      outerContainer: Styles.platformStyles({
        isTablet: {
          flex: 1,
          position: 'relative',
        },
      }),
    } as const)
)

export default Conversation
