import * as React from 'react'
import * as Kb from '../../../../common-adapters'
import * as Styles from '../../../../styles'
import * as ChatTypes from '../../../../constants/types/chat2'
import * as TeamTypes from '../../../../constants/types/teams'
import {Avatars, TeamAvatar} from '../../../avatars'
import {TeamsSubscriberMountOnly} from '../../../../teams/subscriber'

export type ConvProps = {
  fullname: string
  teamType: ChatTypes.TeamType
  teamname: string
  teamID: TeamTypes.TeamID
  ignored: boolean
  muted: boolean
  participants: Array<string>
}

export type Props = {
  attachTo?: () => React.Component<any> | null
  badgeSubscribe: boolean
  canAddPeople: boolean
  channelname?: string
  convProps?: ConvProps
  floatingMenuContainerStyle?: Styles.StylesCrossPlatform
  hasHeader: boolean
  isInChannel: boolean
  isSmallTeam: boolean
  manageChannelsSubtitle: string
  manageChannelsTitle: string
  memberCount: number
  teamname?: string
  visible: boolean
  onAddPeople: () => void
  onBlockConv: () => void
  onHidden: () => void
  onInvite: () => void
  onJoinChannel: () => void
  onLeaveChannel: () => void
  onLeaveTeam: () => void
  onHideConv: () => void
  onMuteConv: (muted: boolean) => void
  onUnhideConv: () => void
  onManageChannels: () => void
  onViewTeam: () => void
  participantsCount: number
}

type AdhocHeaderProps = {
  fullname: string
  isMuted: boolean
  participants: Array<string>
}

const AdhocHeader = (props: AdhocHeaderProps) => (
  <Kb.Box2 direction="horizontal" style={styles.headerContainer}>
    <Avatars
      backgroundColor={Styles.globalColors.white}
      isHovered={false}
      isLocked={false}
      isMuted={props.isMuted}
      isSelected={false}
      participants={props.participants}
      singleSize={32}
    />
    <Kb.Box2 alignItems="flex-start" direction="vertical">
      <Kb.ConnectedUsernames
        colorFollowing={true}
        commaColor={Styles.globalColors.black_50}
        inline={false}
        skipSelf={props.participants.length > 1}
        containerStyle={styles.maybeLongText}
        type="BodyBig"
        underline={false}
        usernames={props.participants}
        onUsernameClicked="profile"
      />
      {!!props.fullname && <Kb.Text type="BodySmall">{props.fullname}</Kb.Text>}
    </Kb.Box2>
  </Kb.Box2>
)

type TeamHeaderProps = {
  isMuted: boolean
  memberCount: number
  teamname: string
  onViewTeam: () => void
}
const TeamHeader = (props: TeamHeaderProps) => {
  return (
    <Kb.Box2 alignItems="center" direction="horizontal" style={styles.headerContainer}>
      <TeamAvatar
        teamname={props.teamname}
        isMuted={props.isMuted}
        isSelected={false}
        isHovered={false}
        size={32}
      />
      <Kb.Box2 direction="horizontal" style={styles.teamText}>
        <Kb.Text type="BodySemibold" style={styles.maybeLongText} onClick={props.onViewTeam}>
          {props.teamname}
        </Kb.Text>
        <Kb.Meta
          backgroundColor={Styles.globalColors.blueGrey}
          color={Styles.globalColors.black_50}
          icon="iconfont-people"
          iconColor={Styles.globalColors.black_20}
          title={props.memberCount}
        />
      </Kb.Box2>
    </Kb.Box2>
  )
}

class InfoPanelMenu extends React.Component<Props> {
  render() {
    const props = this.props
    const isGeneralChannel = !!(props.channelname && props.channelname === 'general')
    const addPeopleItems = [
      {
        icon: 'iconfont-mention',
        onClick: props.onAddPeople,
        style: {borderTopWidth: 0},
        subTitle: 'Keybase, Twitter, etc.',
        title: 'Add someone by username',
      },
      {
        icon: 'iconfont-contact-book',
        onClick: props.onInvite,
        title: Styles.isMobile ? 'Add someone from address book' : 'Add someone by email',
      },
    ]
    const channelHeader = [
      {
        unWrapped: true,
        view: (
          <Kb.Box2 direction="horizontal" fullHeight={true} fullWidth={true} style={styles.channelHeader}>
            <Kb.Text lineClamp={1} type="Body" style={styles.channelName}>
              # <Kb.Text type="BodyBold">{props.channelname}</Kb.Text>
            </Kb.Text>
            <Kb.Meta
              backgroundColor={Styles.globalColors.blueGrey}
              color={Styles.globalColors.black_50}
              icon="iconfont-people"
              iconColor={Styles.globalColors.black_20}
              title={props.participantsCount}
            />
          </Kb.Box2>
        ),
      },
    ]
    const channelItem = props.isSmallTeam
      ? {
          icon: 'iconfont-hash',
          onClick: props.onManageChannels,
          subTitle: props.manageChannelsSubtitle,
          title: props.manageChannelsTitle,
        }
      : {
          icon: 'iconfont-hash',
          isBadged: props.badgeSubscribe,
          onClick: props.onManageChannels,
          title: props.manageChannelsTitle,
        }
    const teamHeader = [
      {
        unWrapped: true,
        view: (
          <Kb.Box2
            direction="horizontal"
            fullHeight={true}
            fullWidth={true}
            style={Styles.collapseStyles([styles.channelHeader, styles.teamHeader])}
          >
            <Kb.Box2 direction="horizontal" gap="tiny">
              <Kb.Avatar teamname={props.teamname} size={16} />
              <Kb.Text type="BodyBold">{props.teamname}</Kb.Text>
            </Kb.Box2>
            <Kb.Meta
              backgroundColor={Styles.globalColors.blueGrey}
              color={Styles.globalColors.black_50}
              icon="iconfont-people"
              iconColor={Styles.globalColors.black_20}
              title={props.memberCount}
            />
          </Kb.Box2>
        ),
      },
    ]

    const isAdhoc =
      (props.isSmallTeam && !props.convProps) || !!(props.convProps && props.convProps.teamType === 'adhoc')
    const items: Kb.MenuItems = (isAdhoc
      ? [
          this.muteItem(),
          this.hideItem(),
          {danger: true, icon: 'iconfont-block-user', onClick: props.onBlockConv, title: 'Block'},
        ]
      : [
          ...(!props.isSmallTeam && !props.hasHeader ? channelHeader : []),
          this.muteItem(),
          this.hideItem(),
          ...(!props.isSmallTeam && !props.isInChannel && !isGeneralChannel && !props.hasHeader
            ? [{onClick: props.onJoinChannel, title: 'Join channel'}]
            : []),
          ...(!props.isSmallTeam && props.isInChannel && !isGeneralChannel && !props.hasHeader
            ? [{danger: true, icon: 'iconfont-leave', onClick: props.onLeaveChannel, title: 'Leave channel'}]
            : []),
          ...(!props.isSmallTeam && !props.hasHeader ? teamHeader : []),
          channelItem,
          {
            icon: 'iconfont-people',
            onClick: props.onViewTeam,
            style: {borderTopWidth: 0},
            title: 'Team info',
          },
          ...(props.canAddPeople ? addPeopleItems : []),
          {danger: true, icon: 'iconfont-leave', onClick: props.onLeaveTeam, title: 'Leave team'},
        ]
    ).reduce<Kb.MenuItems>((arr, i) => {
      i && arr.push(i as Kb.MenuItem)
      return arr
    }, [])

    const header = {
      title: 'header',
      view: props.hasHeader ? (
        isAdhoc && props.convProps ? (
          <AdhocHeader
            isMuted={props.convProps.muted}
            fullname={props.convProps.fullname}
            participants={props.convProps.participants}
          />
        ) : props.teamname ? (
          <TeamHeader
            isMuted={
              props.convProps === null || props.convProps === undefined ? false : props.convProps.muted
            }
            teamname={props.teamname}
            memberCount={props.memberCount}
            onViewTeam={props.onViewTeam}
          />
        ) : null
      ) : null,
    }

    return (
      <>
        {props.visible && <TeamsSubscriberMountOnly />}
        <Kb.FloatingMenu
          attachTo={props.attachTo}
          containerStyle={props.floatingMenuContainerStyle}
          visible={props.visible}
          items={items}
          header={header}
          onHidden={props.onHidden}
          position="bottom left"
          closeOnSelect={true}
        />
      </>
    )
  }

  hideItem() {
    if (this.props.convProps == null) {
      return null
    }
    const convProps = this.props.convProps
    if (convProps.teamType === 'adhoc' || convProps.teamType === 'small') {
      if (convProps.ignored) {
        return {
          icon: 'iconfont-unhide',
          onClick: this.props.onUnhideConv,
          style: {borderTopWidth: 0},
          title: 'Unhide conversation',
        }
      } else {
        return {
          icon: 'iconfont-hide',
          onClick: this.props.onHideConv,
          style: {borderTopWidth: 0},
          title: 'Hide until next message',
        }
      }
    } else {
      return null
    }
  }

  muteItem() {
    if (this.props.convProps == null || !this.props.isInChannel) {
      return null
    }
    const convProps = this.props.convProps
    const title = convProps.muted ? 'Unmute' : 'Mute'
    return {
      icon: 'iconfont-shh',
      iconIsVisible: true,
      onClick: () => this.props.onMuteConv(!convProps.muted),
      title,
    }
  }
}

const styles = Styles.styleSheetCreate(
  () =>
    ({
      badge: Styles.platformStyles({
        common: {
          backgroundColor: Styles.globalColors.blue,
          borderRadius: 6,
          height: 8,
          margin: 6,
          width: 8,
        },
        isElectron: {
          margin: 4,
          marginTop: 5,
          position: 'absolute',
          right: Styles.globalMargins.tiny,
        },
      }),
      channelHeader: {
        ...Styles.padding(Styles.globalMargins.xsmall, Styles.globalMargins.small),
        backgroundColor: Styles.globalColors.blueGreyLight,
        justifyContent: 'space-between',
        marginTop: -Styles.globalMargins.tiny,
      },
      channelName: Styles.platformStyles({
        isElectron: {wordBreak: 'break-all'},
      }),
      headerAvatar: Styles.platformStyles({
        isElectron: {
          marginBottom: 2,
        },
        isMobile: {
          marginBottom: 4,
        },
      }),
      headerContainer: Styles.platformStyles({
        isElectron: {
          ...Styles.padding(
            Styles.globalMargins.small,
            Styles.globalMargins.small,
            Styles.globalMargins.xsmall
          ),
          width: '100%', // don't expand if text is long
        },
        isMobile: {paddingBottom: 24, paddingTop: 40},
      }),
      maybeLongText: Styles.platformStyles({
        isElectron: {
          wordBreak: 'break-word',
        } as const,
      }),
      muteAction: {
        ...Styles.globalStyles.flexBoxRow,
        alignItems: 'center',
      },
      noTopborder: {
        borderTopWidth: 0,
      },
      teamHeader: Styles.platformStyles({
        isElectron: {
          borderTop: `1px solid ${Styles.globalColors.black_10}`,
          marginTop: Styles.globalMargins.tiny,
        },
      }),
      teamText: {
        flex: 1,
        justifyContent: 'space-between',
      },
      text: Styles.platformStyles({
        isMobile: {
          color: Styles.globalColors.blueDark,
        },
      }),
    } as const)
)

export {InfoPanelMenu}
